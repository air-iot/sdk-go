package task

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

type App interface {
	Start(task Task)
	GetCron() *cron.Cron
}

type Task interface {
	Start(App) error
	Stop(App) error
}

func init() {
	viper.SetDefault("log.level", "INFO")
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
	*cron.Cron
}

// NewApp 创建DG
func NewApp() App {
	//var logLevel = viper.GetString("log.level")
	a := new(app)
	//a.Logger = logger.NewLogger(logLevel)
	a.Cron = cron.New(cron.WithSeconds(), cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger)))
	return a
}

// Start 开始服务
func (p *app) Start(task Task) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	if err := task.Start(p); err != nil {
		panic(err)
	}
	p.Cron.Start()
	//p.Logger.Infoln("启动服务")
	sig := <-ch
	close(ch)
	if err := task.Stop(p); err != nil {
		log.Println("任务停止,", err.Error())
	}
	p.Cron.Stop()
	log.Println("关闭服务,", sig)
	os.Exit(0)
}

func (p *app) GetCron() *cron.Cron {
	return p.Cron
}
