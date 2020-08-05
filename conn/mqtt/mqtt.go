package mqtt

import (
	"fmt"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

type Mqtt struct {
	MQTT.Client
}

func NewMqtt(host string, port int, username, password string) (*Mqtt, error) {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", host, port))
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetConnectTimeout(time.Second * 20)
	opts.SetKeepAlive(time.Second * 60)
	opts.SetProtocolVersion(4)
	opts.SetConnectionLostHandler(func(client MQTT.Client, e error) {
		if e != nil {
			logrus.Errorf("消息队列连接错误,%s", e.Error())
		}
	})
	opts.SetOrderMatters(false)
	// Start the connection
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("消息队列连接错误,%s", token.Error())
	}
	return &Mqtt{Client: client}, nil
}

func (p *Mqtt) Close() {
	p.Client.Disconnect(250)
}

// Send 发送消息
func (p *Mqtt) Send(topic, msg string) error {
	if token := p.Client.Publish(topic, 0, false, msg); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Message 订阅并接收消息
func (p *Mqtt) Message(topic string, handler func(client MQTT.Client, message MQTT.Message)) error {
	if token := p.Client.Subscribe(topic, 0, func(client MQTT.Client, message MQTT.Message) {
		handler(client, message)
	}); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
