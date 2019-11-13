package multiple

import (
	"bytes"
	"errors"
	"sync"
	"time"
)

const (
	MAXID uint32 = 0xFFFFFFFF
)

var (
	PROTOCOL   = []byte{'0', 'u', 'c', 'p'}
	PHEAD      = len(PROTOCOL)
	HEADER_LEN = PHEAD + 4 + 2 + 2
	MAXMTU     = 1500
	UCPMTU     = 12 // MAXMTU - 20 - 8 - HEADER_LEN

	pool = sync.Pool{
		New: func() interface{} {
			return make([]byte, MAXMTU)
		},
	}
)

type PacketFragment struct {
	Id      uint32
	Count   uint16
	FragSeq uint16 // optional
	Data    []byte
}

// type PacketMgr uint32

func (p *PacketFragment) Marshall() []byte {
	buf := bytes.NewBuffer(PROTOCOL)
	buf.Write(Uint32ToBytes(p.Id))
	buf.Write(Uint16ToBytes(p.Count))
	if p.Count > 1 {
		buf.Write(Uint16ToBytes(p.FragSeq))
	}
	buf.Write(p.Data)
	return buf.Bytes()
}

func (p *PacketFragment) Unmarshal(m []byte) error {
	if len(m) < HEADER_LEN-2 {
		return errors.New("too short")
	}

	if !bytes.Equal(m[:PHEAD], PROTOCOL) {
		return errors.New("wrong protocol")
	}

	p.Id = BytesToUint32(m[PHEAD : PHEAD+4])
	p.Count = BytesToUint16(m[PHEAD+4 : PHEAD+6])
	var data []byte
	if p.Count > 1 {
		p.FragSeq = BytesToUint16(m[PHEAD+6 : PHEAD+8])
		data = m[PHEAD+8:]
	} else {
		data = m[PHEAD+6:]
	}
	p.Data = pool.Get().([]byte)
	copy(p.Data, data)
	p.Data = p.Data[:len(data)]
	return nil
}

type Packet struct {
	m        map[uint16]*PacketFragment
	lasttime time.Time
	ready    bool
	data     []byte

	Id    uint32
	Count uint16
}

func NewPacket(f *PacketFragment) *Packet {
	p := new(Packet)
	p.Id = f.Id
	p.Count = f.Count
	p.lasttime = time.Now()
	if f.Count <= 1 {
		p.ready = true
		p.data = f.Data
	} else {
		p.m = make(map[uint16]*PacketFragment)
		p.m[f.FragSeq] = f
	}
	return p
}

func (p *Packet) Put(f *PacketFragment) bool {
	if p.ready || p.Id != f.Id || p.Count != f.Count {
		return false
	}

	_, ok := p.m[f.FragSeq]
	if ok {
		return false
	}

	p.m[f.FragSeq] = f
	p.lasttime = time.Now()
	if len(p.m) == int(p.Count) {
		p.data = pool.Get().([]byte)[:0]
		for _, f := range p.m {
			p.data = append(p.data, f.Data...)
			pool.Put(f.Data)
		}
		p.ready = true
	}

	return true
}

func (p *Packet) Ready() bool {
	return p.ready
}

func (p *Packet) ReadyData() []byte {
	return p.data
}

func (p *Packet) Timeouted(to time.Duration) bool {
	return time.Now().After(p.lasttime.Add(to))
}

func (p *Packet) ReleaseMem() {
	if p.ready {
		p.ready = false
		pool.Put(p.data)
	} else {
		for _, f := range p.m {
			pool.Put(f.Data)
		}
	}
	p.data = nil
	p.m = make(map[uint16]*PacketFragment)
}

type MultipleEncoder uint32

// \0ucp(4) packetid(4) count(2) seq(2) length(4) data(length)
func (c *MultipleEncoder) Encode(data []byte) (datas [][]byte) {
	// return m

	datas = Frag(data, UCPMTU)
	for i, data := range datas {
		*c++
		if *c == 0 || uint32(*c) == MAXID { // 0 和 MAXID 只用作界限值
			*c = 1
		}

		var p PacketFragment
		p.Id = uint32(*c)
		p.Count = uint16(len(data))
		p.FragSeq = uint16(i)
		p.Data = data
		datas[i] = p.Marshall()
	}
	return
}

type MultipleDecoder uint32

func (c *MultipleDecoder) Decode(m []byte) *PacketFragment {
	// return m
	var p PacketFragment
	err := p.Unmarshal(m)
	if err != nil {
		return nil
	}
	return &p
}
