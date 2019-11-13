package multiple

import (
	"net"
	"testing"
)

func TestCopy(t *testing.T) {
	dst := make([]byte, 0, 5)
	src := []byte("0123456789")
	t.Log(copy(dst, src))

	dst = dst[:3]
	t.Log(copy(dst, src))

	dst = make([]byte, 15)
	t.Log(copy(dst, src))
}

func TestMap(t *testing.T) {
	m := make(map[int]string)
	t.Log(m)

	f, ok := m[3]
	t.Log(m, f, ok)

	f, ok = m[3]
	t.Log(m, f, ok)

	m[3] = "zt"
	f, ok = m[3]
	t.Log(m, f, ok)
}

func TestIP(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", ":8888")
	ip := addr.IP
	t.Log(addr, ip, addr.Zone, err)
	t.Log(ip.IsGlobalUnicast())
	t.Log(ip.IsInterfaceLocalMulticast())
	t.Log(ip.IsLinkLocalMulticast())
	t.Log(ip.IsLinkLocalUnicast())
	t.Log(ip.IsLoopback())
	t.Log(ip.IsMulticast())
	t.Log(ip.IsUnspecified())
	t.Log(ip.DefaultMask())
}
