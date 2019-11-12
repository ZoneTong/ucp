package multiple

import (
	"io"
	"net"
	"time"
)

type Worker struct {
	conns   []net.Conn
	Network string
	Addrs   []string
	done    chan bool
	sendque chan []byte
	recvque chan []byte
}

var _ io.ReadWriteCloser = new(Worker)

func (w *Worker) Write(d []byte) (n int, err error) {
	return w.Send(d)
}

func (w *Worker) Read(d []byte) (n int, err error) {
	return w.Recv(d)
}

type connReturn struct {
	n   int
	err error
}

func (w *Worker) Send(p []byte) (n int, err error) {
	r := make(chan connReturn)
	for _, conn := range w.conns {
		go func(conn net.Conn) {
			n, err := conn.Write(p)
			r <- connReturn{n, err}
		}(conn)
	}

	var rr connReturn
	for range w.conns {
		rr = <-r
		if rr.err == nil {
			return rr.n, rr.err
		}
	}

	return rr.n, rr.err
}

func (w *Worker) Recv(p []byte) (n int, err error) {
	r := make(chan connReturn)
	for _, conn := range w.conns {
		go func(conn net.Conn) {
			n, err := conn.Read(p)
			r <- connReturn{n, err}
		}(conn)
	}

	var rr connReturn
	for range w.conns {
		rr = <-r
		if rr.err == nil {
			return rr.n, rr.err
		}
	}
	return
}

func (w *Worker) Close() error {
	close(w.done)
	return nil
}

func (w *Worker) Start() error {

	// for _, addr := range worker.Addrs {
	// 	var conn io.ReadWriteCloser
	// 	// conn,err:=GetConn(remote)

	// 	w.conns = append(w.conns, conn)
	// }

	// inputs := multiple.CopiedReader(w, len(w.Addrs), w.done)
	// outputs := multiple.MergedWriter(w, len(w.Addrs), w.done)
	// for i, input := range inputs {
	// 	go func(i int) {
	// 		n, err := io.Copy(w.conns[i], inputs[i])
	// 		if err != nil {
	// 			log.Println(err)
	// 		}
	// 	}()

	// 	go func(i int){
	// 		n,err:=io.Copy(outputs[i],w.conns[i])
	// 		if err!=nil{
	// 			log.Println(err)
	// 		}
	// 	}
	// }
	return nil
}

func NewMultipleWorker(network string, addrs []string, timeout time.Duration) (io.ReadWriteCloser, error) {
	worker := &Worker{
		Network: network,
		Addrs:   addrs,
		done:    make(chan bool),
	}

	for _, addr := range worker.Addrs {
		conn, err := net.DialTimeout(network, addr, timeout)
		if err != nil {
			return nil, err
		}

		worker.conns = append(worker.conns, conn)
	}

	return worker, nil
}
