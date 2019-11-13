package common

import (
	"encoding/binary"
	"net"
	"strings"
)

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

func IsSubIP(parent, son string) bool {
	if parent == son {
		return true
	}

	ph, pp, _ := net.SplitHostPort(parent)
	sh, sp, _ := net.SplitHostPort(son)
	if pp != sp {
		return false
	}

	var ips []string
	if ph == "" || ph == "[::]" {
		addrs, _ := net.InterfaceAddrs()
		for _, addr := range addrs {
			ip := addr.String()
			idx := strings.Index(ip, "/")
			if idx > -1 {
				ip = ip[:idx]
			}
			ips = append(ips, ip)
		}
	}

	// log.Println("IsSubIP", ips)
	for _, ip := range ips {
		if ip == (sh) {
			return true
		}
	}

	return false
}
