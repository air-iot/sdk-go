package sdk

import (
	"errors"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var (
	//ErrClosed 连接池已经关闭Error
	ErrClosed = errors.New("pool is closed")
)

// Pool 基本方法
type Pool interface {
	Get() (MQTT.Client, error)

	Put(MQTT.Client) error

	Close(MQTT.Client) error

	Release()

	Len() int
}