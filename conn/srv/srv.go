package srv

import (
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"

	"github.com/air-iot/sdk-go/conn/mqtt"
	"github.com/air-iot/sdk-go/conn/rabbit"
)

var DataAction = "mqtt"

func DefaultRealtimeDataHandler(handler func(topic string, payload []byte)) error {
	switch viper.GetString("data.action") {
	case "rabbit":
		return NewRabbitService(viper.GetString("service.name"), "data").Consume(rabbit.RoutingKey+"#", handler)
	default:
		return NewMqttService().Consume(mqtt.Topic+"#", handler)
	}
}

type MQService interface {
	Publish(topic string, payload []byte) error
	Consume(topic string, handler func(topic string, payload []byte)) error
}

type PublishFunc func(topic string, msg []byte)

type mqttService struct{}

func NewMqttService() MQService {
	return new(mqttService)
}

func (*mqttService) Publish(topic string, payload []byte) error {
	if token := mqtt.Client.Publish(topic, 0, false, string(payload)); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (*mqttService) Consume(topic string, handler func(topic string, payload []byte)) error {
	if token := mqtt.Client.Subscribe(topic, 0, func(client MQTT.Client, message MQTT.Message) {
		handler(message.Topic(), message.Payload())
	}); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

type rabbitService struct {
	queue    string
	exchange string
}

func NewRabbitService(queue, exchange string) MQService {
	return &rabbitService{
		queue,
		exchange,
	}
}

func NewRabbitEnvService() MQService {
	return &rabbitService{
		viper.GetString("service.queue"),
		"data",
	}
}

func DefaultRabbitService() MQService {
	return &rabbitService{viper.GetString("service.name"), "data"}
}

func (p *rabbitService) newQueue() (amqp.Queue, error) {
	return rabbit.Channel.QueueDeclare(
		p.queue, // name
		true,    // durable
		true,    // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
}

func (p *rabbitService) newExchange() error {
	return rabbit.Channel.ExchangeDeclare(
		p.exchange, // name
		"topic",    // type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
}

func (p *rabbitService) Publish(topic string, payload []byte) error {
	return rabbit.Channel.Publish(
		p.exchange, // exchange
		topic,      // routing key
		false,      // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Transient,
			ContentType:  "text/plain",
			Body:         payload,
		})
}

func (p *rabbitService) Consume(topic string, handler func(topic string, payload []byte)) error {
	q, err := p.newQueue()
	if err != nil {
		return err
	}
	err = p.newExchange()
	if err != nil {
		return err
	}
	err = rabbit.Channel.QueueBind(
		q.Name,     // queue name
		topic,      // routing key
		p.exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}
	err = rabbit.Channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return err
	}
	messages, err := rabbit.Channel.Consume(
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
			//if err := d.Ack(false); err != nil {}
		}
	}()
	return nil
}
