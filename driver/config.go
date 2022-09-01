package driver

import (
	"github.com/air-iot/sdk-go/v4/conn/mq"
	"github.com/air-iot/sdk-go/v4/driver/grpc"
	"github.com/air-iot/sdk-go/v4/logger"
)

// C 全局配置(需要先执行MustLoad，否则拿不到配置)
var C = new(Config)

type Config struct {
	ServiceID string `json:"serviceId" yaml:"serviceId" mapstructure:"serviceId"`
	Project   string `json:"project" yaml:"project" mapstructure:"project"`
	Driver    struct {
		ID   string `json:"id" yaml:"id"`
		Name string `json:"name" yaml:"name"`
	} `json:"driver" yaml:"driver"`
	DriverGrpc grpc.Config `json:"driverGrpc" yaml:"driverGrpc"`
	Log        logger.Log  `json:"log" yaml:"log"`
	MQ         mq.Config   `json:"mq" yaml:"mq"`
}
