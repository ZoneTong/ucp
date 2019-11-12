package multiple

import (
	"fmt"
	"io"
	"net"
)

type pConn struct {
	remote net.Addr
	net.PacketConn
}

func (c *pConn) Read(data []byte) (n int, err error) {
	n, _, err = c.PacketConn.ReadFrom(data)
	return
}

func (c *pConn) Write(data []byte) (n int, err error) {
	return c.PacketConn.WriteTo(data, c.remote)
}

// func ()

func GetConn(host, remoteHost string, port, rport int) (io.ReadWriteCloser, error) {
	// conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(host), Port: port})
	conn, err := net.ListenPacket("udp", fmt.Sprintf("%v:%v", host, port))
	if err != nil {
		return nil, err
	}

	raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", remoteHost, rport))
	if err != nil {
		return nil, err
	}

	return &pConn{remote: raddr, PacketConn: conn}, nil
}
