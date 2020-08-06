/**
 * @Author: ZhangQiang
 * @Description:
 * @File:  app
 * @Version: 1.0.0
 * @Date: 2020/8/6 10:51
 */
package service

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/air-iot/sdk-go/logger"
)

type App interface {
	Start(service Service)
	GetLogger() *logrus.Logger
	GetHttpServer() *echo.Echo
}

type Service interface {
	Start(App) error
	Stop(App) error
}

func init() {
	viper.SetDefault("log.level", "INFO")
	viper.SetDefault("server.port", 8080)

	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./etc/")
	if err := viper.ReadInConfig(); err != nil {
		log.Println("读取配置,", err.Error())
		os.Exit(1)
	}
}

// 任务服务
type app struct {
	*logrus.Logger
	*echo.Echo
}

// NewDG 创建DG
func NewApp() App {
	var logLevel = viper.GetString("log.level")
	a := new(app)
	e := echo.New()
	e.Use(mw.Recover())
	a.Echo = echo.New()
	a.Logger = logger.NewLogger(logLevel)
	return a
}

// Start 开始服务
func (p *app) Start(service Service) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	if err := service.Start(p); err != nil {
		panic(err)
	}
	go func() {
		if err := p.Echo.Start(net.JoinHostPort("", viper.GetString("server.port"))); err != nil {
			p.Logger.Errorln("启动http服务,", err)
			os.Exit(1)
		}
	}()
	p.Logger.Infoln("启动服务")
	sig := <-ch
	close(ch)
	if err := service.Stop(p); err != nil {
		p.Logger.Errorln("任务停止,", err.Error())
	}
	p.stop()
	p.Logger.Infoln("关闭服务,", sig)
	os.Exit(0)
}

func (p *app) stop() {
	if err := p.Echo.Close(); err != nil {
		p.Logger.Errorln("关闭http服务,", err)
	}
}

// GetLogger 获取日志
func (p *app) GetLogger() *logrus.Logger {
	return p.Logger
}

func (p *app) GetHttpServer() *echo.Echo {
	return p.Echo
}
