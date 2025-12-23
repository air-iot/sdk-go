package mq

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/air-iot/logger"
)

type mqtt struct {
	lock      sync.RWMutex
	client    MQTT.Client
	callbacks []Callback
}

// MQTTConfig mqtt配置参数
type MQTTConfig struct {
	Schema          string     `json:"schema"`
	Host            string     `json:"host" yaml:"host"`
	Port            int        `json:"port" yaml:"port"`
	Username        string     `json:"username" yaml:"username"`
	Password        string     `json:"password" yaml:"password"`
	KeepAlive       uint       `json:"keepAlive" yaml:"keepAlive" default:"60"`
	ConnectTimeout  uint       `json:"connectTimeout" yaml:"connectTimeout" default:"20"`
	ProtocolVersion uint       `json:"protocolVersion" yaml:"protocolVersion" default:"4"`
	Order           bool       `json:"order" yaml:"order" default:"false"`
	ClientIdPrefix  string     `json:"clientIdPrefix" yaml:"clientIdPrefix"`
	TLSConfig       *TlsConfig `json:"tlsConfig" yaml:"tlsConfig"`
}

type TlsConfig struct {
	InsecureSkipVerify bool     `json:"insecureSkipVerify" yaml:"insecureSkipVerify"`
	CipherSuites       []string `json:"cipherSuites" yaml:"cipherSuites"`
}

func (a MQTTConfig) DNS() string {
	schema := "tcp"
	if a.Schema != "" {
		schema = a.Schema
	}
	return fmt.Sprintf("%s://%s:%d", schema, a.Host, a.Port)
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
	opts.SetOrderMatters(cfg.Order)
	if cfg.ClientIdPrefix != "" {
		opts.SetClientID(fmt.Sprintf("%s_%s", cfg.ClientIdPrefix, primitive.NewObjectID().Hex()))
	}
	if cfg.TLSConfig != nil {
		if cfg.TLSConfig.InsecureSkipVerify {
			tlsConfig := &tls.Config{InsecureSkipVerify: cfg.TLSConfig.InsecureSkipVerify}
			if cfg.TLSConfig.CipherSuites != nil && len(cfg.TLSConfig.CipherSuites) > 0 {
				tlsConfig.CipherSuites = parseCipherSuites(cfg.TLSConfig.CipherSuites)
			}
			opts.SetTLSConfig(tlsConfig)
		} else if cfg.TLSConfig.CipherSuites != nil && len(cfg.TLSConfig.CipherSuites) > 0 {
			tlsConfig := &tls.Config{}
			tlsConfig.CipherSuites = parseCipherSuites(cfg.TLSConfig.CipherSuites)
			opts.SetTLSConfig(tlsConfig)
		}
	}
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

func parseCipherSuites(names []string) []uint16 {
	var ids []uint16

	// 获取 Go 支持的所有套件：包括标准的和为了兼容性保留的“不安全”套件
	allSuites := append(tls.CipherSuites(), tls.InsecureCipherSuites()...)

	for _, name := range names {
		name = strings.TrimSpace(name)
		for _, s := range allSuites {
			if s.Name == name {
				ids = append(ids, s.ID)
				break
			}
		}
	}
	return ids
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
