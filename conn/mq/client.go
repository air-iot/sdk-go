package mq

import (
	"fmt"
	"strings"
)

const (
	Mqtt   string = "MQTT"
	Rabbit string = "RABBIT"
)

type Config struct {
	Type   string         `json:"type" yaml:"type"`
	MQTT   MQTTConfig     `json:"mqtt" yaml:"mqtt"`
	Rabbit RabbitMQConfig `json:"rabbit" yaml:"rabbit"`
}

type Handler func(topic string, topicSplit []string, payload []byte)

// NewMQ 创建消息队列
func NewMQ(cfg Config) (MQ, func(), error) {
	switch strings.ToUpper(cfg.Type) {
	case Rabbit:
		return NewRabbitClient(cfg.Rabbit)
	case Mqtt:
		return NewMQTTClient(cfg.MQTT)
	default:
		return nil, nil, fmt.Errorf("未知mq类型")
	}
}
