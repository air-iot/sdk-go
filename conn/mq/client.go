package mq

import (
	"fmt"
	"strings"
)

const (
	Mqtt   string = "MQTT"
	Rabbit string = "RABBIT"
	Kafka  string = "KAFKA"
)

type Config struct {
	MQType   string
	MQTT     MQTTConfig
	RabbitMQ RabbitMQConfig
	Kafka    KafkaConfig
}

// NewMQ 创建消息队列
func NewMQ(cfg Config) (MQ, func(), error) {
	switch strings.ToUpper(cfg.MQType) {
	case Rabbit:
		return NewRabbitClient(cfg.RabbitMQ)
	case Mqtt:
		return NewMQTTClient(cfg.MQTT)
	case Kafka:
		return NewKafkaClient(cfg.Kafka)
	default:
		return nil, nil, fmt.Errorf("未知mq类型")
	}
}
