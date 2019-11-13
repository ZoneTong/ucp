package multiple

import (
	"errors"
	"io"
	"log"
	"strings"
	"time"
)

type UserMgr struct {
	conns []localConn

	// Network string
	// Addrs   []string

	Listens []string

	dialTimeout, writeTimeout time.Duration
	readTimeout               time.Duration

	encoder MultipleEncoder
	done    chan bool

	// packet 上层调用send发送的完整数据
	// fragment 底层对packet的分片, 按maxmtu调整
	recvque  chan []byte
	readyque chan *Packet
	reader   *packetReader
}

var (
	_ io.ReadWriteCloser = new(UserMgr)
)

type connReturn struct {
	n   int
	err error
}

func (w *UserMgr) Write(p []byte) (int, error) {
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
			return len(p), rr.err
		} else {
			errs = append(errs, rr.err.Error())
		}
	}

	return 0, errors.New(strings.Join(errs, "|"))
}

func (w *UserMgr) writeConn(conn localConn, datas [][]byte, r chan<- connReturn) {
	var total, n int
	var err error
	if w.writeTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(w.writeTimeout))
	}
	for _, data := range datas {
		n, err = conn.Write(data)
		total += n
		if err != nil {
			break
		}
	}
	r <- connReturn{total, err}
}

func (w *UserMgr) readStart() {
	w.reader.SetReading(true)
	go w.reader.refreshReadyMinid()
}

func (w *UserMgr) readEnd() {
	w.reader.SetReading(false)
}

func (w *UserMgr) Read(p []byte) (n int, err error) {
	var timer = time.NewTimer(0)
	timer.Stop()
	if w.readTimeout > 0 {
		timer.Reset(w.readTimeout)
	}

	w.readStart()
	defer w.readEnd()

	var packet *Packet
	select {
	case <-timer.C:
		return 0, errors.New("read timeout")
	case packet = <-w.readyque:
	}

	n = copy(p, packet.ReadyData())
	packet.ReleaseMem()
	return
}

func (w *UserMgr) recvLoop(conn localConn) {
	for {
		data := pool.Get().([]byte)
		n, err := conn.Read(data)
		if err != nil {
			log.Println("recvLoop err", n, err)
			return
		}

		log.Println("recvLoop before")
		w.recvque <- data[:n]
		log.Println("recvLoop after")
	}
}

func (w *UserMgr) Close() error {
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

func (w *UserMgr) establishConnection(conns []localConn) (err error) {
	w.conns = conns

	// go w.timerClearTimeouReadingPacket()
	go w.reader.decodeRecvDataLoop()
	for _, conn := range w.conns {
		go w.recvLoop(conn)
	}
	return nil
}

func NewUserMgr(conns []localConn, timeout []string) (io.ReadWriteCloser, error) {
	w := &UserMgr{
		// Network: network,
		// Addrs:   addrs,
		done:     make(chan bool),
		recvque:  make(chan []byte),
		readyque: make(chan *Packet),
	}
	buffertimeout, _ := time.ParseDuration(timeout[3])
	w.reader = NewPacketReader(w.recvque, w.readyque, w.done, buffertimeout)

	w.readTimeout, _ = time.ParseDuration(timeout[2])
	w.writeTimeout, _ = time.ParseDuration(timeout[1])
	w.dialTimeout, _ = time.ParseDuration(timeout[0])

	err := w.establishConnection(conns)
	if err != nil {
		return nil, err
	}
	return w, nil
}
