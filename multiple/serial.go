package multiple

import "encoding/binary"

func Uint32ToBytes(v uint32) (bs []byte) {
	bs = make([]byte, 4)
	binary.BigEndian.PutUint32(bs, v)
	return
}

func BytesToUint32(bs []byte) uint32 {
	return binary.BigEndian.Uint32(bs)
}

func Uint16ToBytes(v uint16) (bs []byte) {
	bs = make([]byte, 2)
	binary.BigEndian.PutUint16(bs, v)
	return
}

func BytesToUint16(bs []byte) uint16 {
	return binary.BigEndian.Uint16(bs)
}

// mtu=1500 whole IP layer
// -20 ip header
// -8  udp header
// = 1472
func Frag(data []byte, mtu int) (payload [][]byte) {
	total := len(data)
	for total > mtu {
		payload = append(payload, data[:mtu])
		data = data[mtu:]
		total -= mtu
	}
	if total > 0 {
		payload = append(payload, data)
	}

	return
}
