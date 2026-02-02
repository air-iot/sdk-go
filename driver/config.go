package driver

import (
	apiConfig "github.com/air-iot/api-client-go/v4/config"
	"github.com/air-iot/logger"
	"github.com/air-iot/sdk-go/v4/conn/mq"
	"github.com/air-iot/sdk-go/v4/driver/grpc"
	"github.com/air-iot/sdk-go/v4/etcd"
)

// Cfg 全局配置(需要先执行MustLoad，否则拿不到配置)
var Cfg = new(Config)

type Mode string

const (
	LocalMode  Mode = "local"  // 本地模式：启动 HTTP 服务器，处理 data.json
	NormalMode Mode = "normal" // 正常模式：连接 gRPC，不处理 data.json
)

type Config struct {
	ServiceID string `json:"serviceId" yaml:"serviceId" mapstructure:"serviceId"`
	GroupID   string `json:"groupId" yaml:"groupId" mapstructure:"groupId"`
	Project   string `json:"project" yaml:"project" mapstructure:"project"`
	Driver    struct {
		ID   string `json:"id" yaml:"id"`
		Name string `json:"name" yaml:"name"`
	} `json:"driver" yaml:"driver"`
	Mode       Mode          `json:"mode" yaml:"mode" mapstructure:"mode"`
	DriverGrpc grpc.Config   `json:"driverGrpc" yaml:"driverGrpc"`
	Log        logger.Config `json:"log" yaml:"log"`
	MQ         mq.Config     `json:"mq" yaml:"mq"`
	Pprof      struct {
		Enable bool   `json:"enable" yaml:"enable"`
		Host   string `json:"host" yaml:"host"`
		Port   string `json:"port" yaml:"port"`
	} `json:"pprof" yaml:"pprof"`
	HTTP struct {
		Host string `json:"host" yaml:"host"`
		Port string `json:"port" yaml:"port"`
	} `json:"http" yaml:"http"`
	DataConfig string           `json:"dataConfig" yaml:"dataConfig"` // data.json 文件路径
	EtcdConfig string           `json:"etcdConfig" yaml:"etcdConfig"`
	Etcd       etcd.Config      `json:"etcd" yaml:"etcd"`
	API        apiConfig.Config `json:"api" yaml:"api"`
}
