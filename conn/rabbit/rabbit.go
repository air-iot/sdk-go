package rabbit

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

var Channel *amqp.Channel
var conn *amqp.Connection
var RoutingKey = "data."
var Queue string

func Init() {
	var (
		host     = viper.GetString("rabbit.host")
		port     = viper.GetInt("rabbit.port")
		username = viper.GetString("rabbit.username")
		password = viper.GetString("rabbit.password")
		vhost    = viper.GetString("rabbit.vhost")
	)
	Queue = viper.GetString("rabbit.queue")
	RoutingKey = viper.GetString("rabbit.routingKey")

	var err error
	conn, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/%s", username, password, host, port, vhost))
	if err != nil {
		logrus.Panic(err)
	}
	Channel, err = conn.Channel()
	if err != nil {
		logrus.Panic(err)
	}
}

func Close() {
	if Channel != nil && !conn.IsClosed() {
		if err := Channel.Close(); err != nil {
			logrus.Errorln("关闭RabbitMQ channel错误", err.Error())
		}
	}
	if conn != nil && !conn.IsClosed() {
		if err := conn.Close(); err != nil {
			logrus.Errorln("关闭RabbitMQ connect错误", err.Error())
		}
	}
}
