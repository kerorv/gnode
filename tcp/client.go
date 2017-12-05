package tcp

import (
	"net"
	"time"
)

type Client struct {
	s       *session
	codec   Codec
	handler uint32
}

func NewClient(codec Codec, handler uint32) *Client {
	return &Client{
		s:       nil,
		codec:   codec,
		handler: handler,
	}
}

func (c *Client) Start(address string) bool {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return false
	}

	c.s = newSession(conn, c.handler, c.codec)
	c.s.start()
	return true
}

func (c *Client) StartTimeout(address string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}

	c.s = newSession(conn, c.handler, c.codec)
	c.s.start()
	return true
}

func (c *Client) Stop() {
	c.s.stop()
}
