package mq

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/air-iot/sdk-go/v4/logger"
)

type mqtt struct {
	lock      sync.RWMutex
	client    MQTT.Client
	callbacks []Callback
}

// MQTTConfig mqtt配置参数
type MQTTConfig struct {
	Host            string `json:"host" yaml:"host"`
	Port            int    `json:"port" yaml:"port"`
	Username        string `json:"username" yaml:"username"`
	Password        string `json:"password" yaml:"password"`
	KeepAlive       uint   `json:"keepAlive" yaml:"keepAlive" default:"60"`
	ConnectTimeout  uint   `json:"connectTimeout" yaml:"connectTimeout" default:"20"`
	ProtocolVersion uint   `json:"protocolVersion" yaml:"protocolVersion" default:"4"`
}

func (a MQTTConfig) DNS() string {
	return fmt.Sprintf("tcp://%s:%d", a.Host, a.Port)
}

const TOPICSEPWITHMQTT = "/"

func NewMQTT(cli MQTT.Client) MQ {
	m := new(mqtt)
	m.client = cli
	return m
}

// NewMQTTClient 创建MQTT消息队列
func NewMQTTClient(cfg MQTTConfig) (MQ, func(), error) {
	mqCli := new(mqtt)
	mqCli.callbacks = make([]Callback, 0)
	opts := MQTT.NewClientOptions()
	opts.AddBroker(cfg.DNS())
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)
	opts.SetConnectTimeout(time.Second * 20)
	opts.SetKeepAlive(time.Second * 60)
	opts.SetProtocolVersion(4)
	opts.SetConnectionLostHandler(func(client MQTT.Client, e error) {
		if e != nil {
			logger.Errorf("MQTT Lost错误: %s", e.Error())
			mqCli.lost()
		}
	})
	opts.SetOrderMatters(false)
	opts.SetOnConnectHandler(func(client MQTT.Client) {
		logger.Infof("MQTT 已连接")
		mqCli.connect()
	})
	// Start the connection
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, nil, token.Error()
	}
	cleanFunc := func() {
		client.Disconnect(250)
	}
	mqCli.client = client
	return mqCli, cleanFunc, nil
}

func (p *mqtt) Callback(cb Callback) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.callbacks = append(p.callbacks, cb)
	return
}

func (p *mqtt) lost() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for _, cb := range p.callbacks {
		if err := cb.Lost(p); err != nil {
			logger.Fatalf("lost callback err, %s", err)
		}
	}
	return
}

func (p *mqtt) connect() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for _, cb := range p.callbacks {
		if err := cb.Connect(p); err != nil {
			logger.Fatalf("connect callback err, %s", err)
		}
	}
	return
}

func (p *mqtt) Publish(ctx context.Context, topicParams []string, payload []byte) error {
	topic := strings.Join(topicParams, TOPICSEPWITHMQTT)
	if token := p.client.Publish(topic, 0, false, string(payload)); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (p *mqtt) Consume(ctx context.Context, topicParams []string, splitN int, handler Handler) error {
	topic := strings.Join(topicParams, TOPICSEPWITHMQTT)
	if token := p.client.Subscribe(topic, 0, func(client MQTT.Client, message MQTT.Message) {
		handler(message.Topic(), strings.SplitN(message.Topic(), TOPICSEPWITHMQTT, splitN), message.Payload())
	}); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (p *mqtt) UnSubscription(ctx context.Context, topicParams []string) error {
	topic := strings.Join(topicParams, TOPICSEPWITHMQTT)
	if token := p.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
