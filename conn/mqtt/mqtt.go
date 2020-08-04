package mqtt

import (
	"fmt"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Client MQTT.Client
var Topic = "data/"

func Init() {
	opts := MQTT.NewClientOptions()
	var host = viper.GetString("mqtt.host")
	var port = viper.GetInt("mqtt.port")
	var username = viper.GetString("mqtt.username")
	var password = viper.GetString("mqtt.password")
	Topic = viper.GetString("mqtt.topic")
	var clientID = viper.GetString("mqtt.clientId")
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", host, port))
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetConnectTimeout(time.Second * 20)
	opts.SetKeepAlive(time.Second * 60)
	opts.SetProtocolVersion(4)
	opts.SetClientID(clientID)
	//opts.SetConnectionLostHandler()
	opts.SetConnectionLostHandler(func(client MQTT.Client, e error) {
		if e != nil {
			logrus.Panic(e)
		}
	})
	opts.SetOrderMatters(false)
	// Start the connection
	Client = MQTT.NewClient(opts)
	if token := Client.Connect(); token.Wait() && token.Error() != nil {
		logrus.Panic(token.Error())
	}
}

func Close() {
	if Client != nil {
		Client.Disconnect(250)
	}
}

// SendMsg 发送消息
func Send(topic string, b []byte) error {
	if token := Client.Publish(topic, 0, false, string(b)); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// RecMsg 接收消息
func Rec(topic string, handler func(client MQTT.Client, message MQTT.Message)) error {
	if token := Client.Subscribe(topic, 0, func(client MQTT.Client, message MQTT.Message) {
		handler(client, message)
	}); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
