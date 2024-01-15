package flow

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/air-iot/logger"
)

type App interface {
	Start(flow Flow)
}

// app 数据采集类
type app struct {
	stopped bool
	cli     *Client
	clean   func()
}

func init() {
	// 设置随机数种子
	rand.Seed(time.Now().Unix())
	runtime.GOMAXPROCS(runtime.NumCPU())
	cfgPath := pflag.String("config", "./etc/", "配置文件")
	pflag.Parse()
	viper.SetDefault("log.level", 4)
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("flowEngine.host", "flow-engine")
	viper.SetDefault("flowEngine.port", 2333)
	viper.SetDefault("flow.timeout", 600)
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(*cfgPath)
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		log.Fatalln("读取命令行参数错误,", err.Error())
	}
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln("读取配置错误,", err.Error())
	}

	if err := viper.Unmarshal(Cfg); err != nil {
		log.Fatalln("配置解析错误: ", err.Error())
	}
}

// NewApp 创建App
func NewApp() App {
	a := new(app)
	if Cfg.Flow.Mode == "" || Cfg.Flow.Name == "" {
		panic("流程节点name和模式不能为空")
	}
	Cfg.Log.Syslog.ServiceName = Cfg.Flow.Name
	logger.InitLogger(Cfg.Log)
	logger.Debugf("配置: %+v", *Cfg)
	a.clean = func() {}
	return a
}

// Start 开始服务
func (a *app) Start(flow Flow) {
	a.stopped = false
	cli := Client{}
	a.cli = cli.Start(a, flow)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	sig := <-ch
	close(ch)
	cli.Stop()
	logger.Debugf("关闭服务: 信号=%v", sig)
	os.Exit(0)
}
