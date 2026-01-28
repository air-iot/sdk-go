package mq

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/air-iot/logger"
)

// Local 本地 MQ 实现，用于没有 MQ server 时驱动的开发和测试
// Publish 操作会将消息输出到日志
// Consume 操作不会真正消费消息，仅记录日志
type local struct {
	lock       sync.RWMutex
	callbacks  []Callback
	subscribed map[string]bool // 记录订阅的主题
}

// LocalConfig 本地 MQ 配置参数
type LocalConfig struct {
	// LogPublish 是否记录发布消息到日志，默认 true
	LogPublish bool `json:"logPublish" yaml:"logPublish"`
	// LogConsume 是否记录消费消息到日志，默认 true
	LogConsume bool `json:"logConsume" yaml:"logConsume"`
	// ShowPayload 是否显示消息内容，默认 false（只显示长度）
	ShowPayload bool `json:"showPayload" yaml:"showPayload"`
}

const TOPICSEPWITHLOCAL = "/"

// NewLocal 创建本地 MQ 实例
func NewLocal(cfg LocalConfig) (MQ, func(), error) {
	m := &local{
		callbacks:  make([]Callback, 0),
		subscribed: make(map[string]bool),
	}

	// 设置默认值
	if !cfg.LogPublish && !cfg.LogConsume {
		cfg.LogPublish = true
		cfg.LogConsume = true
	}

	logger.Infof("使用本地MQ模式: logPublish=%v, logConsume=%v, showPayload=%v",
		cfg.LogPublish, cfg.LogConsume, cfg.ShowPayload)

	// 触发连接回调
	m.connect()

	cleanFunc := func() {
		logger.Infof("本地MQ已关闭")
	}

	return m, cleanFunc, nil
}

func (p *local) Callback(cb Callback) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.callbacks = append(p.callbacks, cb)
}

func (p *local) connect() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for _, cb := range p.callbacks {
		if err := cb.Connect(p); err != nil {
			logger.Errorf("connect callback err, %s", err)
		}
	}
}

func (p *local) lost() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for _, cb := range p.callbacks {
		if err := cb.Lost(p); err != nil {
			logger.Errorf("lost callback err, %s", err)
		}
	}
}

// Publish 发布消息到本地（输出到日志）
func (p *local) Publish(ctx context.Context, topicParams []string, payload []byte) error {
	topic := strings.Join(topicParams, TOPICSEPWITHLOCAL)

	payloadInfo := fmt.Sprintf("len=%d", len(payload))
	if len(payload) <= 100 {
		payloadInfo = fmt.Sprintf("data=%s", string(payload))
	} else {
		payloadInfo = fmt.Sprintf("len=%d, data=%s...", len(payload), string(payload[:100]))
	}

	logger.Infof("[MQ-Publish] topic=%s, %s", topic, payloadInfo)
	return nil
}

// Consume 订阅消息（本地模式不会真正消费消息）
func (p *local) Consume(ctx context.Context, topicParams []string, splitN int, handler Handler) error {
	topic := strings.Join(topicParams, TOPICSEPWITHLOCAL)

	p.lock.Lock()
	if p.subscribed[topic] {
		p.lock.Unlock()
		logger.Debugf("[MQ-Consume] 主题已订阅: %s", topic)
		return nil
	}
	p.subscribed[topic] = true
	p.lock.Unlock()

	logger.Infof("[MQ-Consume] 订阅主题: %s (本地模式不会接收消息)", topic)
	return nil
}

// UnSubscription 取消订阅
func (p *local) UnSubscription(ctx context.Context, topicParams []string) error {
	topic := strings.Join(topicParams, TOPICSEPWITHLOCAL)

	p.lock.Lock()
	delete(p.subscribed, topic)
	p.lock.Unlock()

	logger.Infof("[MQ-Unsubscribe] 取消订阅主题: %s", topic)
	return nil
}
