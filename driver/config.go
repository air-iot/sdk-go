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

type Config struct {
	ServiceID string `json:"serviceId" yaml:"serviceId" mapstructure:"serviceId"`
	GroupID   string `json:"groupId" yaml:"groupId" mapstructure:"groupId"`
	Project   string `json:"project" yaml:"project" mapstructure:"project"`
	Driver    struct {
		ID   string `json:"id" yaml:"id"`
		Name string `json:"name" yaml:"name"`
	} `json:"driver" yaml:"driver"`
	DriverGrpc grpc.Config   `json:"driverGrpc" yaml:"driverGrpc"`
	Log        logger.Config `json:"log" yaml:"log"`
	MQ         mq.Config     `json:"mq" yaml:"mq"`
	Pprof      struct {
		Enable bool   `json:"enable" yaml:"enable"`
		Host   string `json:"host" yaml:"host"`
		Port   string `json:"port" yaml:"port"`
	} `json:"pprof" yaml:"pprof"`
	HTTP struct {
		Enable bool   `json:"enable" yaml:"enable"`
		Mode   string `json:"debug" yaml:"debug"`
		Host   string `json:"host" yaml:"host"`
		Port   string `json:"port" yaml:"port"`
	} `json:"http" yaml:"http"`
	Datafile struct {
		Enable bool   `json:"enable" yaml:"enable"`
		Path   string `json:"path" yaml:"path"` // data.json 文件路径
	}
	License string `json:"license" yaml:"license"`
	//DataConfig string           `json:"dataConfig" yaml:"dataConfig"` // data.json 文件路径
	EtcdConfig string           `json:"etcdConfig" yaml:"etcdConfig"`
	Etcd       etcd.Config      `json:"etcd" yaml:"etcd"`
	API        apiConfig.Config `json:"api" yaml:"api"`
}
