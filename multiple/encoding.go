package multiple

import (
	"bytes"
	"errors"
)

var (
	PROTOCOL   = []byte{0x0, 'u', 'c', 'p'}
	PHEAD      = len(PROTOCOL)
	HEADER_LEN = PHEAD + 4 + 2 + 2
	MAXMTU     = 1500
	UCPMTU     = MAXMTU - 20 - 8 - HEADER_LEN
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
	if p.Count > 1 {
		p.FragSeq = BytesToUint16(m[PHEAD+6 : PHEAD+8])
		p.Data = m[PHEAD+8:]
	} else {
		p.Data = m[PHEAD+6:]
	}
	return nil
}

type MultipleEncoder uint32

// \0ucp(4) packetid(4) count(2) seq(2) length(4) data(length)
func (c *MultipleEncoder) Encode(data []byte) (datas [][]byte) {
	// return m

	datas = Frag(data, UCPMTU)
	for i, data := range datas {
		*c++
		if *c == 0 {
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
