package multiple

import (
	"errors"
	"io"
	"log"
	"strings"
	"sync"
	"time"
)

type Worker struct {
	conns []localConn

	// Network string
	// Addrs   []string

	Listens []string

	dialTimeout, writeTimeout  time.Duration
	readTimeout, buffertimeout time.Duration

	encoder MultipleEncoder
	decoder MultipleDecoder
	done    chan bool

	// packet 上层调用send发送的完整数据
	// fragment 底层对packet的分片, 按maxmtu调整
	recvque      chan []byte
	recv_lock    sync.Mutex
	recv_packets map[uint32]*Packet
	ready_minid  uint32
	readyque     chan *Packet

	reading_lock sync.Mutex
	reading      bool
}

type reader struct {
	done     chan bool
	recvque  <-chan []byte
	readyque chan<- *Packet

	recv_lock    sync.Mutex
	recv_packets map[uint32]*Packet
	ready_minid  uint32
	reading_lock sync.Mutex
	reading      bool
}

var (
	_ io.ReadWriteCloser = new(Worker)
)

type connReturn struct {
	n   int
	err error
}

func (w *Worker) Write(p []byte) (int, error) {
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

func (w *Worker) writeConn(conn localConn, datas [][]byte, r chan<- connReturn) {
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

func (w *Worker) SetReading(b bool) {
	w.reading_lock.Lock()
	w.reading = b
	w.reading_lock.Unlock()
}

func (w *Worker) Reading() bool {
	w.reading_lock.Lock()
	b := w.reading
	w.reading_lock.Unlock()
	return b
}

func (w *Worker) Read(p []byte) (n int, err error) {
	var timer = time.NewTimer(0)
	timer.Stop()
	if w.readTimeout > 0 {
		timer.Reset(w.readTimeout)
	}

	w.SetReading(true)
	defer w.SetReading(false)
	go w.refreshReadyMinid()

	var packet *Packet
	select {
	case <-timer.C:
		return 0, errors.New("read timeout")
	case packet = <-w.readyque:
	}

	n = copy(p, packet.ReadyData())
	packet.ReleaseMem()
	// go func() {
	// 	select {
	// 	case w.recvedsig <- true:
	// 	}
	// }()
	return
}

func (w *Worker) recvLoop(conn localConn) {
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

// 定时清除超时packet
func (w *Worker) timerClearTimeouReadingPacket() {
	timer := time.NewTimer(w.buffertimeout)
	for {
		select {
		case <-w.done:
			timer.Stop()
			return
		// case <-w.recvedsig:
		case <-timer.C:
		}

		w.clearTimeoutPackets(w.buffertimeout)
		timer.Reset(w.buffertimeout)
	}
}

// 带锁,故需用gorouting调用
func (w *Worker) refreshReadyMinid() {
	w.recv_lock.Lock()
	defer w.recv_lock.Unlock()
	log.Println("refreshReadyMinid start", w.ready_minid)
	defer log.Println("refreshReadyMinid end ready_minid", w.ready_minid)
	var min_id uint32 = MAXID
	for id := range w.recv_packets {
		if id < min_id {
			min_id = id
		}
	}

	if min_id == MAXID {
		w.ready_minid = 0
		return
	}

	p, ok := w.recv_packets[min_id]
	if ok && p.Ready() {
		w.ready_minid = min_id
		// log.Println("refreshReadyMinid Reading  before", w.ready_minid)
		if w.Reading() {
			// log.Println("refreshReadyMinid Reading  after", w.ready_minid)
			log.Println("before send readyque", w.recv_packets[min_id])
			delete(w.recv_packets, min_id)
			go func() {
				w.readyque <- p
			}()
			return
		}
		// log.Println("refreshReadyMinid Reading  not", w.ready_minid)
	}
}

func (w *Worker) clearTimeoutPackets(to time.Duration) {
	w.recv_lock.Lock()
	defer w.recv_lock.Unlock()
	log.Println("clearTimeoutPackets")
	for id, packet := range w.recv_packets {
		if packet.Timeouted(to) {
			log.Println("clearTimeoutPackets Timeouted", packet.Id, packet.ReadyData())
			packet.ReleaseMem()
			delete(w.recv_packets, id)
		}
	}
	go w.refreshReadyMinid()
}

func (w *Worker) decodeRecvDataLoop() {
	var p *PacketFragment
	for {
		select {
		case <-w.done:
			return

		case d := <-w.recvque:
			log.Println("decodeRecvDataLoop before", d)
			p = w.decoder.Decode(d)
			pool.Put(d)
		}

		if p == nil {
			log.Println("decode fail")
			continue
		}

		w.recvPacketFragment(p)
	}

}

func (w *Worker) recvPacketFragment(f *PacketFragment) {
	w.recv_lock.Lock()
	defer w.recv_lock.Unlock()
	log.Println("recvPacketFragment", f)
	if f.Id <= w.ready_minid {
		pool.Put(f.Data)
		return
	}

	p, ok := w.recv_packets[f.Id]
	if !ok {
		p = NewPacket(f)
		w.recv_packets[p.Id] = p
	} else {
		if !p.Put(f) {
			pool.Put(f.Data)
		}
	}

	log.Println("recvPacketFragment packet", p)
	if p.Ready() {
		go w.refreshReadyMinid()
	}
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

func (w *Worker) establishConnection(conns []localConn) (err error) {
	w.conns = conns

	w.recv_packets = make(map[uint32]*Packet)
	w.recvque = make(chan []byte)
	w.readyque = make(chan *Packet)

	// go w.timerClearTimeouReadingPacket()
	go w.decodeRecvDataLoop()
	for _, conn := range w.conns {
		go w.recvLoop(conn)
	}
	return nil
}

func NewMultipleWorker(conns []localConn, timeout []string) (io.ReadWriteCloser, error) {
	worker := &Worker{
		// Network: network,
		// Addrs:   addrs,
		done: make(chan bool),
	}

	worker.buffertimeout, _ = time.ParseDuration(timeout[3])
	if worker.buffertimeout == 0 {
		worker.buffertimeout = time.Second * 3
	}
	worker.readTimeout, _ = time.ParseDuration(timeout[2])
	worker.writeTimeout, _ = time.ParseDuration(timeout[1])
	worker.dialTimeout, _ = time.ParseDuration(timeout[0])

	err := worker.establishConnection(conns)
	if err != nil {
		return nil, err
	}
	return worker, nil
}
