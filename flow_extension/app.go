package flow_extionsion

import "C"
import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/air-iot/logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type App interface {
	Start(ext Extension)
}

// app 数据采集类
type app struct {
	*logrus.Logger
	stopped bool
	cli     *Client
	clean   func()
}

func init() {
	// 设置随机数种子
	runtime.GOMAXPROCS(runtime.NumCPU())
	cfgPath := pflag.String("config", "./etc/", "配置文件")
	pflag.Parse()
	viper.SetDefault("log.level", 4)
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("flowEngine.host", "flow-engine")
	viper.SetDefault("flowEngine.port", 2333)
	viper.SetDefault("extension.timeout", 600)
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	log.Println("配置文件路径", *cfgPath)
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
	if Cfg.Extension.Id == "" || Cfg.Extension.Name == "" {
		panic("流程扩展服务 id 和 name 不能为空")
	}
	C.Log.Syslog.ServiceName = Cfg.Extension.Id
	logger.InitLogger(C.Log)
	logger.Debugf("配置: %+v", *Cfg)
	a.clean = func() {}
	return a
}

// Start 开始服务
func (a *app) Start(ext Extension) {
	a.stopped = false
	cli := Client{}
	a.cli = cli.Start(a, ext)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	sig := <-ch
	close(ch)
	cli.Stop()
	logger.Debugln("关闭服务,", sig)
	os.Exit(0)
}
