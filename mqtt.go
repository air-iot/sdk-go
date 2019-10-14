package sdk

import (
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// Init 初始化
func Init(addr, username, password string) (MQTT.Client, error) {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(addr)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetConnectTimeout(time.Second * 20)
	opts.SetKeepAlive(time.Second * 60)
	opts.SetProtocolVersion(4)
	//opts.SetConnectionLostHandler()
	// Start the connection
	cc := MQTT.NewClient(opts)
	if token := cc.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return cc, nil
}

// SendMsg 发送消息
func Send(client MQTT.Client, topic string, b []byte) error {
	if token := client.Publish(topic, 0, false, string(b)); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// RecMsg 接收消息
func Rec(client MQTT.Client, topic string, handler func(client MQTT.Client, message MQTT.Message)) error {
	if token := client.Subscribe(topic, 0, func(client MQTT.Client, message MQTT.Message) {
		handler(client, message)
	}); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
