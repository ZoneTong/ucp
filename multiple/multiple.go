package multiple

import (
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type Framentmap struct {
	m     map[uint16][]byte
	first time.Time
}

type Worker struct {
	conns []net.Conn

	Network string
	Addrs   []string
	encoder MultipleEncoder
	decoder MultipleDecoder

	dialTimeout, writeTimeout, readTimeout time.Duration

	// packet 上层调用send发送的完整数据
	// fragment 底层对packet的分片, 按maxmtu调整
	recvque        chan []byte
	reading_lock   sync.Mutex
	reading_packet map[uint32]*Framentmap
	ready_lock     sync.Mutex
	ready_packet   map[uint32][]byte
	recved_id      uint32
	done           chan bool
}

var (
	_    io.ReadWriteCloser = new(Worker)
	pool                    = sync.Pool{
		New: func() interface{} {
			return make([]byte, MAXMTU)
		},
	}
)

type connReturn struct {
	n   int
	err error
}

func (w *Worker) Write(p []byte) (n int, err error) {
	r := make(chan connReturn)
	datas := w.encoder.Encode(p)
	for _, conn := range w.conns {
		go w.writeConn(conn, datas, r)
	}

	var errs []string
	var rr connReturn
	for range w.conns {
		rr = <-r
		if rr.err == nil {
			return rr.n, rr.err
		} else {
			errs = append(errs, err.Error())
		}
	}

	return 0, errors.New(strings.Join(errs, "|"))
}

func (w *Worker) writeConn(conn net.Conn, datas [][]byte, r chan<- connReturn) {
	var total, n int
	var err error
	for _, data := range datas {
		n, err = conn.Write(data)
		total += n
		if err != nil {
			break
		}
	}
	r <- connReturn{total, err}
}

func (w *Worker) Read(p []byte) (n int, err error) {
	r := make(chan connReturn)
	for _, conn := range w.conns {
		go func(conn net.Conn) {
			n, err := conn.Read(p)
			r <- connReturn{n, err}
		}(conn)
	}

	var rr connReturn
	var errs []string
	for range w.conns {
		rr = <-r
		if rr.err == nil {
			return rr.n, rr.err
		} else {
			errs = append(errs, err.Error())
		}
	}
	return 0, errors.New(strings.Join(errs, "|"))
}

func (w *Worker) recvLoop(conn net.Conn, que chan []byte) {
	for {
		data := pool.Get().([]byte)
		n, err := conn.Read(data)
		if err != nil {
			return
		}

		que <- data[:n]
	}
}

func (w *Worker) timerClearReadingPacket() {

}

func (w *Worker) mergePacket() {
	var p *PacketFragment
	for {
		select {
		case <-w.done:
			return

		case d := <-w.recvque:
			p = w.decoder.Decode(d)
			if p == nil {
				continue
			}

			w.reading_lock.Lock()
			if p.Id > w.recved_id {
				if p.Count <= 1 { // 没有分片
					w.ready_lock.Lock()
					if _, ok := w.ready_packet[p.Id]; !ok {
						w.ready_packet[p.Id] = p.Data
					}
					w.ready_lock.Unlock()
				} else { // 等待聚合
					m, ok := w.reading_packet[p.Id]
					if !ok {
						m = new(Framentmap)
						m.m = make(map[uint16][]byte)
						m.first = time.Now()
						w.reading_packet[p.Id] = m
					}

					_, ok = m.m[p.FragSeq]
					if !ok {
						m.m[p.FragSeq] = p.Data
					}
				}
			}
			w.reading_lock.Unlock()
		}
	}
}

func (w *Worker) readConn(conn net.Conn, datas [][]byte, r chan<- connReturn) {
	var total, n int
	var err error
	for _, data := range datas {
		n, err = conn.Read(data)
		total += n
		if err != nil {
			break
		}
	}
	r <- connReturn{total, err}
}

func (w *Worker) Close() error {
	var errs []string
	for _, conn := range w.conns {
		err := conn.Close()
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) != 0 {
		return errors.New(strings.Join(errs, "|"))
	}
	close(w.done)
	return nil
}

func (w *Worker) establishConnection() (err error) {
	var conn net.Conn
	for _, addr := range w.Addrs {
		if w.dialTimeout > 0 {
			conn, err = net.DialTimeout(w.Network, addr, w.dialTimeout)
		} else {
			conn, err = net.Dial(w.Network, addr)
		}

		if err != nil {
			return
		}

		w.conns = append(w.conns, conn)
	}
	return nil
}

func NewMultipleWorker(network string, addrs []string, timeout []string) (io.ReadWriteCloser, error) {
	worker := &Worker{
		Network: network,
		Addrs:   addrs,
		done:    make(chan bool),
	}

	switch len(timeout) {
	case 3:
		worker.writeTimeout, _ = time.ParseDuration(timeout[2])
		fallthrough
	case 2:
		worker.readTimeout, _ = time.ParseDuration(timeout[1])
		fallthrough
	case 1:
		worker.dialTimeout, _ = time.ParseDuration(timeout[0])
	}

	err := worker.establishConnection()
	if err != nil {
		return nil, err
	}
	return worker, nil
}
