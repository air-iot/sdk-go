package tcp

import (
	"fmt"
	"net"

	"github.com/air-iot/logger"
	"github.com/pkg/errors"
)

type Conn struct {
	net.Conn
	network string
	address string
	isClose bool
}

func DialTCP(network, host string, port int) (*Conn, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	c, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	conn := new(Conn)
	conn.network = network
	conn.address = address
	conn.Conn = c
	return conn, nil
}

func (c *Conn) Read(b []byte) (n int, err error) {
	if c.isClose {
		return 0, errors.New("连接已主动关闭！")
	}
	n, err = c.Conn.Read(b)
	if err != nil {
		err1 := c.reConnect()
		if err1 != nil {
			return n, err
		}
		return c.Conn.Read(b)
	}
	return n, err
}

func (c *Conn) Write(b []byte) (n int, err error) {
	if c.isClose {
		return -1, errors.New("连接已主动关闭！")
	}
	n, err = c.Conn.Write(b)
	if err != nil {
		err1 := c.reConnect()
		if err1 != nil {
			return n, err
		}
		return c.Conn.Write(b)
	}
	return n, err
}

func (c *Conn) reConnect() error {
	conn, err := net.Dial(c.network, c.address)
	if err != nil {
		return err
	}
	c.Conn = conn
	return nil
}

func (c *Conn) Close() {
	if err := c.Conn.Close(); err != nil {
		logger.Errorf("socket连接关闭失败:%s", err.Error())
	}
	c.isClose = true

}
