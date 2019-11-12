package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	multiple "ucp/multiple"
)

type GlobalConfig struct {
	Users map[string]*UserHandler `json:"users"`
}

func (c *GlobalConfig) Start() error {
	var errs []string
	for _, h := range c.Users {
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
	for _, h := range c.Users {
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
	Network string   `json:"network"`
	Addrs   []string `json:"addrs"`
	Timeout int      `json:"timeout"`
	worker  io.ReadWriteCloser
}

func (c *UserHandler) Key() string {
	return fmt.Sprint(c.Network, c.Addrs)
}

func (c *UserHandler) Start() error {
	w, err := multiple.NewMultipleWorker(c.Network, c.Addrs, time.Duration(c.Timeout)*time.Second)
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

const MTU = 1500

func (c *UserHandler) Recv() ([]byte, error) {
	bs := make([]byte, MTU)
	n, err := c.worker.Read(bs)
	return bs[:n], err
}
