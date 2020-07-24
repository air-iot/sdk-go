package traefik

import "github.com/spf13/viper"

var Host string
var Port int
var Proto = "http"
var Enable = false
var AppKey string
var AppSecret string

func Init() {
	Host = viper.GetString("traefik.host")
	Port = viper.GetInt("traefik.port")
	Enable = viper.GetBool("traefik.enable")
	Proto = viper.GetString("traefik.schema")
	AppKey = viper.GetString("traefik.appKey")
	AppSecret = viper.GetString("traefik.appSecret")
}
