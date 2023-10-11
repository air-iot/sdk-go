package mq

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rabbitmq/amqp091-go"

	"github.com/air-iot/logger"
)

type rabbit struct {
	//queue    string
	//exchange string
	conn *amqp091.Connection
}

// RabbitMQConfig rabbitmq配置参数
type RabbitMQConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	VHost    string `json:"vHost" yaml:"vHost"`
	Exchange string `json:"exchange" yaml:"exchange"`
	Queue    string `json:"queue" yaml:"queue"`
}

func (a RabbitMQConfig) DNS() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		a.Username,
		a.Password,
		a.Host,
		a.Port,
		a.VHost)
}

const TOPICSEPWITHRABBIT = "."

func NewRabbitClient(cfg RabbitMQConfig) (MQ, func(), error) {
	conn, err := amqp091.Dial(cfg.DNS())
	if err != nil {
		return nil, nil, fmt.Errorf("创建AMQP客户端错误: %+v", err)
	}
	m := new(rabbit)
	m.conn = conn
	cleanFunc := func() {
		if err := conn.Close(); err != nil {
			logger.Errorf("rabbitmq close error: %s", err.Error())
		}
	}
	return m, cleanFunc, nil
}

func (p *rabbit) Callback(Callback) {}

func (p *rabbit) NewQueue(channel *amqp091.Channel, queueName string) (*amqp091.Queue, error) {
	queue, err := channel.QueueDeclare(
		queueName, // name
		true,      // durable
		true,      // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, err
	}
	return &queue, nil
}

func (p *rabbit) NewExchange(channel *amqp091.Channel, exchange string) error {
	return channel.ExchangeDeclare(
		exchange, // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
}

func (p *rabbit) Publish(ctx context.Context, topicParams []string, payload []byte) error {
	channel, err := p.conn.Channel()
	if err != nil {
		return err
	}
	exchange, ok := ctx.Value("exchange").(string)
	if !ok {
		return errors.New("context exchange not found")
	}
	topic := strings.Join(topicParams, TOPICSEPWITHRABBIT)
	return channel.Publish(
		exchange, // exchange
		topic,    // routing key
		false,    // mandatory
		false,
		amqp091.Publishing{
			DeliveryMode: amqp091.Transient,
			ContentType:  "text/plain",
			Body:         payload,
		})
}

func (p *rabbit) Consume(ctx context.Context, topicParams []string, splitN int, handler Handler) error {
	channel, err := p.conn.Channel()
	if err != nil {
		return err
	}

	queue, ok := ctx.Value("exchange").(string)
	if !ok {
		return errors.New("context queue not found")
	}

	exchange, ok := ctx.Value("exchange").(string)
	if !ok {
		return errors.New("context exchange not found")
	}
	q, err := p.NewQueue(channel, queue)
	if err != nil {
		return err
	}
	err = p.NewExchange(channel, exchange)
	if err != nil {
		return err
	}
	topic := strings.Join(topicParams, TOPICSEPWITHRABBIT)
	err = channel.QueueBind(
		q.Name,   // queue name
		topic,    // routing key
		exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}
	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return err
	}
	messages, err := channel.Consume(
		q.Name, // queue
		topic,  // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}
	go func() {
		for d := range messages {
			handler(d.RoutingKey, strings.SplitN(d.RoutingKey, TOPICSEPWITHRABBIT, splitN), d.Body)
			//if err := d.Ack(false); err != nil {}
		}
	}()
	return nil
}

func (p *rabbit) UnSubscription(ctx context.Context, topicParams []string) error {
	topic := strings.Join(topicParams, TOPICSEPWITHRABBIT)
	channel, err := p.conn.Channel()
	if err != nil {
		return err
	}
	return channel.Cancel(topic, true)
}
