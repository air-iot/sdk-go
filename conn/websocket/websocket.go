package websocket

import (
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Conn *websocket.Conn

func Init() {
	var (
		scheme = viper.GetString("websocket.scheme")
		host   = viper.GetString("websocket.host")
		port   = viper.GetInt("websocket.port")
		path   = viper.GetString("websocket.path")
	)
	u := url.URL{Scheme: scheme, Host: fmt.Sprintf("%s:%d", host, port), Path: path}
	var err error
	Conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logrus.Panic(err)
	}
}

func Close() {
	if Conn != nil {
		if err := Conn.Close(); err != nil {
			logrus.Errorln("关闭websocket错误", err.Error())
		}
	}
}
