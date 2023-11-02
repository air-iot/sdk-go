package mq

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/air-iot/logger"
)

var _ MQ = new(kafka)

type kafka struct {
	lock                 sync.RWMutex
	config               KafkaConfig
	client               sarama.Client
	producer             sarama.SyncProducer
	consumer             sarama.ConsumerGroup
	callbacks            []Callback
	consumerTopicContext map[string]context.CancelFunc
}

// KafkaConfig mqtt配置参数
type KafkaConfig struct {
	Brokers         []string
	Version         string
	GroupID         string
	ClientID        string
	MaxOpenRequests int
	Balancer        string
	Partition       *int32
	AutoCommit      *bool
}

// NewKafkaClient 创建Kafka消息队列
func NewKafkaClient(cfg KafkaConfig) (mq MQ, clean func(), err error) {
	cli := new(kafka)
	cli.config = cfg
	cli.callbacks = make([]Callback, 0)
	cli.consumerTopicContext = make(map[string]context.CancelFunc)
	client, err := cli.getClient()
	if err != nil {
		return nil, nil, err
	}
	cleanFunc := func() {
		logger.Infof("关闭kafka客户端")
		if err := client.Close(); err != nil {
			logger.Errorf("关闭kafka客户端错误:%v", err)
		}
	}
	cli.client = client
	return cli, cleanFunc, nil
}

func (k *kafka) Publish(_ context.Context, topicParams []string, payload []byte) error {
	if len(topicParams) == 0 {
		return fmt.Errorf("topic为空")
	}
	topic := topicParams[0]
	key := topic
	if len(topicParams) >= 2 {
		key = strings.Join(topicParams[1:], TOPICSEPWITHMQTT)
	}
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(payload),
	}
	if k.config.Partition != nil {
		msg.Partition = *k.config.Partition
	}
	producer, err := k.getProducer()
	if err != nil {
		return err
	}
	_, _, err = producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("发送消息错误:%w", err)
	}
	return nil
}

func (k *kafka) Consume(ctx context.Context, topicParams []string, splitN int, handler Handler) error {
	if len(topicParams) == 0 {
		return fmt.Errorf("topic为空")
	}
	errChan := make(chan error)
	newCtx, newCancel := context.WithCancel(ctx)
	kh := &kafkaHandler{topicParams: topicParams, topicString: strings.Join(topicParams, TOPICSEPWITHMQTT), splitN: splitN, handler: handler, k: k, errChan: errChan, cancel: newCancel}
	go func() {
		select {
		case <-newCtx.Done():
			logger.Infof("订阅数据,发起停止,topic:%+v", topicParams)
		case err := <-errChan:
			if err != nil {
				logger.Errorf("订阅数据错误,收到管道信息,topic:%+v,错误:%v", topicParams, err)
			}
		}
		//close(errChan)
		k.clean()
	}()
	return kh.consume(newCtx)
}

func (k *kafka) UnSubscription(_ context.Context, topicParams []string) error {
	if len(topicParams) == 0 {
		return fmt.Errorf("topic为空")
	}
	//k.cancelFunc(topicParams[0])
	return nil
}

func (k *kafka) Callback(cb Callback) {
	k.lock.Lock()
	defer k.lock.Unlock()
	k.callbacks = append(k.callbacks, cb)
	return
}

func (k *kafka) getConfig() (c *sarama.Config, err error) {
	version := sarama.V0_10_2_0
	if len(k.config.Brokers) == 0 {
		return nil, fmt.Errorf("kafka地址为空")
	}
	if k.config.Version != "" {
		version, err = sarama.ParseKafkaVersion(k.config.Version)
		if err != nil {
			return nil, fmt.Errorf("解析kafka的version错误:%w", err)
		}
	}
	config := sarama.NewConfig()
	config.Version = version
	config.Consumer.Return.Errors = true
	config.Producer.Return.Successes = true
	//config.Consumer.Group.Rebalance.Timeout = 120 * time.Second
	if k.config.MaxOpenRequests != 0 {
		config.Net.MaxOpenRequests = k.config.MaxOpenRequests
	}
	if k.config.AutoCommit != nil {
		config.Consumer.Offsets.AutoCommit.Enable = *k.config.AutoCommit
	}
	switch k.config.Balancer {
	case "Hash":
		config.Producer.Partitioner = sarama.NewHashPartitioner
	case "ReferenceHash":
		config.Producer.Partitioner = sarama.NewReferenceHashPartitioner
	case "CRC32Balancer":
		config.Producer.Partitioner = sarama.NewConsistentCRCHashPartitioner
	case "Murmur2Balancer":
		config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	}
	if k.config.ClientID != "" {
		config.ClientID = k.config.ClientID
	}
	return config, nil
}

func (k *kafka) getClient() (sarama.Client, error) {
	k.lock.Lock()
	defer k.lock.Unlock()
	tempConfig, err := k.getConfig()
	if err != nil {
		return nil, err
	}
	client, err := sarama.NewClient(k.config.Brokers, tempConfig)
	if err != nil {
		return nil, fmt.Errorf("创建kafka客户端错误:%w", err)
	}
	return client, nil
}

func (k *kafka) getProducer() (sarama.SyncProducer, error) {
	if k.client == nil {
		return nil, fmt.Errorf("客户端为空")
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	if k.producer == nil {
		producer, err := sarama.NewSyncProducerFromClient(k.client)
		if err != nil {
			return nil, fmt.Errorf("创建生产者错误:%w", err)
		}
		k.producer = producer
	}
	return k.producer, nil
}

func (k *kafka) clean() {
	time.Sleep(time.Second)
	k.lost()
	k.connect()
}

func (k *kafka) lost() {
	k.lock.Lock()
	defer k.lock.Unlock()
	logger.Infof("lost callback")
	for _, cb := range k.callbacks {
		if err := cb.Lost(k); err != nil {
			logger.Fatalf("lost callback err, %s", err)
		}
	}
	return
}

func (k *kafka) connect() {
	k.lock.Lock()
	defer k.lock.Unlock()
	logger.Infof("connect callback")
	for _, cb := range k.callbacks {
		if err := cb.Connect(k); err != nil {
			logger.Fatalf("connect callback err, %s", err)
		}
	}
	return
}

type kafkaHandler struct {
	splitN      int
	handler     Handler
	k           *kafka
	topicParams []string
	topicString string
	cancel      context.CancelFunc
	errChan     chan error
}

func (h *kafkaHandler) Setup(sess sarama.ConsumerGroupSession) error {
	logger.Infof("kafka handler setup,topic:%+v,MemberID:%s,GenerationID:%d", h.topicParams, sess.MemberID(), sess.GenerationID())
	return nil
}

func (h *kafkaHandler) Cleanup(sess sarama.ConsumerGroupSession) error {
	logger.Infof("kafka handler cleanup,topic:%+v,MemberID:%s,GenerationID:%d", h.topicParams, sess.MemberID(), sess.GenerationID())
	h.cancel()
	return nil
}

func (h *kafkaHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.handlerMessage(sess, claim, msg)
	}
	return nil
}

func (h *kafkaHandler) handlerMessage(sess sarama.ConsumerGroupSession, _ sarama.ConsumerGroupClaim, msg *sarama.ConsumerMessage) {
	if msg == nil {
		return
	}
	defer func() {
		sess.MarkMessage(msg, "")
	}()
	topic := strings.Join([]string{msg.Topic, string(msg.Key)}, TOPICSEPWITHMQTT)
	if h.mqttMatch(h.topicString, topic) {
		h.handler(topic, strings.SplitN(topic, TOPICSEPWITHMQTT, h.splitN), msg.Value)
	}
}

func (h *kafkaHandler) mqttMatch(subscription, topic string) bool {
	subscription = strings.ReplaceAll(subscription, "+", "[^/]+")
	subscription = strings.ReplaceAll(subscription, "#", ".*")
	subscription = "^" + subscription + "$"

	match, _ := regexp.MatchString(subscription, topic)
	return match
}

func (h *kafkaHandler) consume(ctx context.Context) error {
	consumer, err := sarama.NewConsumerGroupFromClient(h.k.config.GroupID, h.k.client)
	if err != nil {
		return fmt.Errorf("创建消费者错误:%w", err)
	}
	topic := h.topicParams[0]
	go func() {
		select {
		case <-ctx.Done():
			logger.Infof("订阅数据,上下文消费错误停止,topic:%s", topic)
			return
		case err := <-consumer.Errors():
			logger.Errorf("订阅数据,收到错误,topic:%s,错误:%v", topic, err)
			h.cancel()
			h.errChan <- err
		}
	}()
	go func() {
		logger.Infof("订阅数据,topic:%s", topic)
		if err := consumer.Consume(ctx, []string{topic}, h); err != nil {
			logger.Errorf("订阅数据,消费错误,topic:%s,错误:%v", topic, err)
			h.errChan <- err
		}
	}()
	return nil
}
