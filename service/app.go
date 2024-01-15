package service

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type App interface {
	Start(service Service)
	GetHttpServer() *gin.Engine
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
	*gin.Engine
	srv *http.Server
}

// NewApp 创建DG
func NewApp() App {
	//var logLevel = viper.GetString("log.level")
	a := new(app)
	g := gin.New()
	g.Use(RecoveryMiddleware())
	a.Engine = g
	//a.Logger = logger.NewLogger(logLevel)
	return a
}

// Start 开始服务
func (p *app) Start(service Service) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	if err := service.Start(p); err != nil {
		panic(err)
	}
	p.srv = &http.Server{
		Addr:    net.JoinHostPort("", viper.GetString("server.port")),
		Handler: p.Engine,
	}
	go func() {
		if err := p.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Println("启动服务")
	sig := <-ch
	close(ch)
	if err := service.Stop(p); err != nil {
		log.Println("任务停止,", err.Error())
	}
	p.stop()
	log.Println("关闭服务,", sig)
	os.Exit(0)
}

func (p *app) stop() {
	if err := p.srv.Close(); err != nil {
		log.Println("关闭http服务,", err)
	}
}

func (p *app) GetHttpServer() *gin.Engine {
	return p.Engine
}
