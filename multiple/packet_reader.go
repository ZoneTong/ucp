package multiple

import (
	"log"
	"sync"
	"time"
)

type packetReader struct {
	done     <-chan bool
	recvque  <-chan []byte
	readyque chan<- *Packet

	buffertimeout time.Duration
	decoder       MultipleDecoder

	recv_lock    sync.Mutex
	recv_packets map[uint32]*Packet
	ready_minid  uint32
	reading_lock sync.Mutex
	reading      bool
}

func (w *packetReader) SetReading(b bool) {
	w.reading_lock.Lock()
	w.reading = b
	w.reading_lock.Unlock()
}

func (w *packetReader) Reading() bool {
	w.reading_lock.Lock()
	b := w.reading
	w.reading_lock.Unlock()
	return b
}

// 定时清除超时packet
func (w *packetReader) timerClearTimeouReadingPacket() {
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
func (w *packetReader) refreshReadyMinid() {
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

func (w *packetReader) clearTimeoutPackets(to time.Duration) {
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

func (w *packetReader) decodeRecvDataLoop() {
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

func (w *packetReader) recvPacketFragment(f *PacketFragment) {
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

func NewPacketReader(recvque <-chan []byte, readyque chan<- *Packet, done <-chan bool,
	buffertimeout time.Duration) *packetReader {
	reader := new(packetReader)
	reader.done = done
	reader.recvque = recvque
	reader.readyque = readyque
	reader.recv_packets = make(map[uint32]*Packet)
	reader.buffertimeout = buffertimeout
	if reader.buffertimeout == 0 {
		reader.buffertimeout = time.Second * 3
	}
	return reader
}
