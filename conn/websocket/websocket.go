package websocket

import (
	"errors"

	"github.com/air-iot/logger"
	"github.com/gorilla/websocket"
)

type Conn struct {
	*websocket.Conn
	url     string
	isClose bool
}

func DialWS(url string) (*Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &Conn{Conn: conn, url: url}, nil
}

func (c *Conn) Read() (n int, b []byte, err error) {
	if c.isClose {
		return -1, nil, errors.New("连接已主动关闭！")
	}
	n, b, err = c.Conn.ReadMessage()
	if err != nil {
		err1 := c.reConnect()
		if err1 != nil {
			return -1, nil, err
		}
		return c.Conn.ReadMessage()
	}
	return n, b, err
}

func (c *Conn) ReadJson(v interface{}) (err error) {
	if c.isClose {
		return errors.New("连接已主动关闭！")
	}
	err = c.Conn.ReadJSON(v)
	if err != nil {
		err1 := c.reConnect()
		if err1 != nil {
			return err
		}
		return c.Conn.ReadJSON(v)
	}
	return err
}

func (c *Conn) Write(messageType int, data []byte) (err error) {
	if c.isClose {
		return errors.New("连接已主动关闭！")
	}
	err = c.Conn.WriteMessage(messageType, data)
	if err != nil {
		err1 := c.reConnect()
		if err1 != nil {
			return err
		}
		return c.Conn.WriteMessage(messageType, data)
	}
	return err
}

func (c *Conn) WriteJson(v interface{}) (err error) {
	if c.isClose {
		return errors.New("连接已主动关闭！")
	}
	err = c.Conn.WriteJSON(v)
	if err != nil {
		err1 := c.reConnect()
		if err1 != nil {
			return err
		}
		return c.Conn.WriteJSON(v)
	}
	return err
}

func (c *Conn) reConnect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return err
	}
	//c.cli.Close()
	c.Conn = conn
	return nil
}

func (c *Conn) Close() {
	if err := c.Conn.Close(); err != nil {
		logger.Errorln("关闭websocket错误", err.Error())
	}
	c.isClose = true
}
