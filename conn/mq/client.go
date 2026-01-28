package mq

import (
	"fmt"
	"strings"
	"time"
)

const (
	Mqtt   string = "MQTT"
	Rabbit string = "RABBIT"
	Kafka  string = "KAFKA"
	Local  string = "LOCAL"
)

type Config struct {
	Type    string         `json:"type" yaml:"type"`
	Timeout time.Duration  `json:"timeout" yaml:"timeout"`
	MQTT    MQTTConfig     `json:"mqtt" yaml:"mqtt"`
	Rabbit  RabbitMQConfig `json:"rabbit" yaml:"rabbit"`
	Kafka   KafkaConfig    `json:"kafka" yaml:"kafka"`
	Local   LocalConfig    `json:"local" yaml:"local"`
}

// NewMQ 创建消息队列
func NewMQ(cfg Config) (MQ, func(), error) {
	switch strings.ToUpper(cfg.Type) {
	case Rabbit:
		return NewRabbitClient(cfg.Rabbit)
	case Mqtt:
		return NewMQTTClient(cfg.MQTT)
	case Kafka:
		return NewKafkaClient(cfg.Kafka)
	case Local:
		return NewLocal(cfg.Local)
	default:
		// 如果未指定类型或类型为空，默认使用本地模式
		if cfg.Type == "" {
			return NewLocal(cfg.Local)
		}
		return nil, nil, fmt.Errorf("未知mq类型: %s", cfg.Type)
	}
}
