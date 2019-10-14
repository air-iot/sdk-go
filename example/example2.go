package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/casic-iot/sdk-go"
)

type TestDriver2 struct{}

func (*TestDriver2) Start(dg *sdk.DG, models []byte) error {
	fmt.Printf("start 长度:%d,数据:%s\n", len(models), string(models))
	return nil
}

func (*TestDriver2) Reload(dg *sdk.DG, models []byte) error {
	fmt.Printf("reload 长度:%d,数据:%s\n", len(models), string(models))
	return nil
}

func (*TestDriver2) Run(dg *sdk.DG, deviceID string, cmd []byte) error {
	fmt.Printf("run 设备:%s,数据:%s", deviceID, string(cmd))
	return nil
}

func main() {
	dg := sdk.NewDG(sdk.ServiceConfig{
		Schema: `{"name":"测试2"}`,
		Consul: &sdk.GCConfig{
			Host: "localhost",
			Port: 8500,
		},
		Service: &sdk.RegistryConfig{
			ID:   "aisenz_tdm",
			Name: "aisenz_tdm_service",
			Tags: []string{"driver.schedule=loadbalance"},
		},
		Driver: &sdk.DriverConfig{
			ID:   "aisenz_tdm",
			Name: "艾森Lora/TDM111",
		},
		Gateway: &sdk.GCConfig{
			Host: "localhost",
			Port: 8082,
		},
		Mqtt: &sdk.EmqttConfig{
			Host:        "localhost",
			Port:        1883,
			Username:    "admin",
			Password:    "public",
		},
	})

	defer func() {
		if err := dg.Stop(); err != nil {
			log.Println("服务停止失败", err)
		} else {
			log.Println("服务停止成功")
		}
	}()

	err := dg.Start(new(TestDriver2))
	if err != nil {
		panic("服务启动失败" + err.Error())
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	select {
	// wait on kill signal
	case sig := <-ch:
		log.Printf("Received signal %s\n", sig)
	}
}
