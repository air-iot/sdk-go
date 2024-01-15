package rabbit

import (
	"fmt"

	"github.com/air-iot/logger"
	"github.com/rabbitmq/amqp091-go"
)

type Amqp struct {
	*amqp091.Connection
}

func NewAmqp(host string, port int, username, password, vhost string) (*Amqp, error) {
	conn, err := amqp091.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/%s", username, password, host, port, vhost))
	if err != nil {
		return nil, fmt.Errorf("创建Amqp连接错误,%s", err.Error())
	}

	return &Amqp{Connection: conn}, nil
}

func (p *Amqp) Close() {
	if !p.Connection.IsClosed() {
		if err := p.Connection.Close(); err != nil {
			logger.Errorln("关闭Amqp连接错误", err.Error())
		}
	}
}

func (p *Amqp) Send(exchange, routerKey string, data []byte) error {

	channel, err := p.Connection.Channel()
	if err != nil {
		return fmt.Errorf("打开Channel错误,%s", err.Error())
	}

	return channel.Publish(exchange, // exchange
		routerKey, // routing key
		false,     // mandatory
		false,
		amqp091.Publishing{
			DeliveryMode: amqp091.Transient,
			ContentType:  "text/plain",
			Body:         data,
		})
}

func (p *Amqp) Message(exchange, routingKey, queue string, handler func(routingKey string, body []byte)) error {
	channel, err := p.Connection.Channel()
	if err != nil {
		return fmt.Errorf("打开Channel错误,%s", err.Error())
	}
	q, err := channel.QueueDeclare(
		queue, // name
		true,  // durable
		true,  // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}
	err = channel.ExchangeDeclare(
		exchange, // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return err
	}
	err = channel.QueueBind(
		q.Name,     // queue name
		routingKey, // routing key
		exchange,   // exchange
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
		"",     // consumer
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
			handler(d.RoutingKey, d.Body)
		}
	}()
	return nil
}
