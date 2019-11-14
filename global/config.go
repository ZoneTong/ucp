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
	Endpoints map[string]*UserConfig `json:"endpoints"`
	// Servers map[string]*UserConfig `json:"servers"`
	Mtu int `json:"mtu"`
}

func (c *GlobalConfig) Start() error {
	var errs []string
	for _, h := range c.Endpoints {
		// fmt.Println(tag, h)
		err := h.Start()
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	// for _, h := range c.Servers {
	// 	// fmt.Println(tag, h)
	// 	err := h.Start()
	// 	if err != nil {
	// 		errs = append(errs, err.Error())
	// 	}
	// }

	if len(errs) != 0 {
		return errors.New(strings.Join(errs, "|"))
	}

	return nil
}

func (c *GlobalConfig) Close() error {
	var errs []string
	for _, h := range c.Endpoints {
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

type UserConfig struct {
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

func (c *UserConfig) Key() string {
	return fmt.Sprint(c.Network, c.Addrs)
}

func (c *UserConfig) Start() error {
	dialTimeout, _ := time.ParseDuration(c.DialTimeout)
	conns, err := multiple.EstablishConnection(c.Network, c.Listen, c.Addrs, dialTimeout)
	if err != nil {
		return err
	}

	w, err := multiple.NewUserMgr(conns, []string{c.DialTimeout, c.WriteTimeout, c.ReadTimeout, c.BufferTimeout})
	if err != nil {
		return err
	}
	c.worker = w
	return nil
}

func (c *UserConfig) Close() error {
	return c.worker.Close()
}

func (c *UserConfig) Send(p []byte) (int, error) {
	return c.worker.Write(p)
}

func (c *UserConfig) Recv(p []byte) (int, error) {
	return c.worker.Read(p)
}
