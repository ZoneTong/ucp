package multiple

import (
	"errors"
	"io"
	"net"
	"strings"
	"time"
)

type localConn interface {
	io.ReadWriteCloser
	LocalAddr() net.Addr
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

type UDPConnection struct {
	*net.UDPConn
	remoteAddrs []net.Addr
}

func (uc *UDPConnection) Read(p []byte) (n int, err error) {
	var raddr net.Addr
	var ok bool
	for !ok {
		n, raddr, err = uc.UDPConn.ReadFromUDP(p)
		// log.Println("UDPConnection raddr", raddr)
		ok = uc.isInRemoteAddrsGroup(raddr)
	}
	return
}

func (uc *UDPConnection) Write(p []byte) (n int, err error) {
	// log.Println("UDPConnection start")
	var errs []string
	var ok bool
	for _, addr := range uc.remoteAddrs {
		// log.Println("UDPConnection write before", addr)
		n1, err1 := uc.UDPConn.WriteTo(p, addr)
		// log.Println("UDPConnection write after", n1, err1)
		if err1 != nil {
			errs = append(errs, err1.Error())
			continue
		}
		ok = true
		n = n1
	}

	if ok {
		return
	} else {
		err = errors.New(strings.Join(errs, "|"))
	}
	return
}

func (uc *UDPConnection) isInRemoteAddrsGroup(addr net.Addr) bool {
	if len(uc.remoteAddrs) == 0 {
		return true
	}

	for _, remote := range uc.remoteAddrs {
		// log.Println(remote, addr)

		if IsSubIP(remote.String(), addr.String()) {
			return true
		}
	}
	return false
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

func EstablishConnection(network, listen string, addrs []string, dialTimeout time.Duration) (conns []localConn, err error) {
	if len(listen) > 0 {
		// for _,addr:=range
		addr, err := net.ResolveUDPAddr(network, listen)
		if err != nil {
			return nil, err
		}
		uconn, err := net.ListenUDP(network, addr)
		if err != nil {
			return nil, err
		}

		var udpaddrs []net.Addr
		for _, addr := range addrs {
			udpaddr, err := net.ResolveUDPAddr(network, addr)
			if err != nil {
				return nil, err
			}
			udpaddrs = append(udpaddrs, udpaddr)
		}

		conns = append(conns, &UDPConnection{UDPConn: uconn, remoteAddrs: udpaddrs})
	} else {
		var conn localConn
		for _, addr := range addrs {
			if dialTimeout > 0 {
				conn, err = net.DialTimeout(network, addr, dialTimeout)
			} else {
				conn, err = net.Dial(network, addr)
			}

			if err != nil {
				return
			}

			conns = append(conns, conn)
		}
	}
	return
}
