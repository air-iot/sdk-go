package driver

import (
	"time"

	"github.com/air-iot/logger"
	"github.com/air-iot/sdk-go/v4/conn/mq"
	"github.com/air-iot/sdk-go/v4/driver/grpc"
)

// Cfg 全局配置(需要先执行MustLoad，否则拿不到配置)
var Cfg = new(Config)

type Config struct {
	ServiceID string `json:"serviceId" yaml:"serviceId" mapstructure:"serviceId"`
	GroupID   string `json:"groupId" yaml:"groupId" mapstructure:"groupId"`
	Project   string `json:"project" yaml:"project" mapstructure:"project"`
	Driver    struct {
		ID      string        `json:"id" yaml:"id"`
		Name    string        `json:"name" yaml:"name"`
		Timeout time.Duration `json:"timeout" yaml:"timeout"`
	} `json:"driver" yaml:"driver"`
	DriverGrpc grpc.Config   `json:"driverGrpc" yaml:"driverGrpc"`
	Log        logger.Config `json:"log" yaml:"log"`
	MQ         mq.Config     `json:"mq" yaml:"mq"`
}
