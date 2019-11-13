package global

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	multiple "ucp/multiple"
)

type GlobalConfig struct {
	Clients map[string]*UserHandler `json:"endpoints"`
	Servers map[string]*UserHandler `json:"servers"`
	Mtu     int                     `json:"mtu"`
}

func (c *GlobalConfig) Start() error {
	var errs []string
	for _, h := range c.Clients {
		// fmt.Println(tag, h)
		err := h.Start()
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	for _, h := range c.Servers {
		// fmt.Println(tag, h)
		err := h.Start()
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) != 0 {
		return errors.New(strings.Join(errs, "|"))
	}

	return nil
}

func (c *GlobalConfig) Close() error {
	var errs []string
	for _, h := range c.Clients {
		err := h.Close()
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) != 0 {
		return errors.New(strings.Join(errs, "|"))
	}
	return nil
}

type UserHandler struct {
	// Tag     string
	Network       string   `json:"network"`
	Addrs         []string `json:"addrs"`
	DialTimeout   string   `json:"dialtimeout"`
	ReadTimeout   string   `json:"readtimeout"`
	WriteTimeout  string   `json:"writetimeout"`
	BufferTimeout string   `json:"buffertimeout"`

	Listen string `json:"listen"`
	worker io.ReadWriteCloser
}

func (c *UserHandler) Key() string {
	return fmt.Sprint(c.Network, c.Addrs)
}

func (c *UserHandler) Start() error {
	dialTimeout, _ := time.ParseDuration(c.DialTimeout)
	conns, err := multiple.EstablishConnection(c.Network, c.Listen, c.Addrs, dialTimeout)
	if err != nil {
		return err
	}

	w, err := multiple.NewMultipleWorker(conns, []string{c.DialTimeout, c.WriteTimeout, c.ReadTimeout, c.BufferTimeout})
	if err != nil {
		return err
	}
	c.worker = w
	return nil
}

func (c *UserHandler) Close() error {
	return c.worker.Close()
}

func (c *UserHandler) Send(p []byte) (int, error) {
	return c.worker.Write(p)
}

func (c *UserHandler) Recv() ([]byte, error) {
	bs := make([]byte, multiple.MAXMTU)
	n, err := c.worker.Read(bs)
	return bs[:n], err
}
