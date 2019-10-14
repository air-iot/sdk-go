# sdk-go
支持用户通过go开发驱动，并接入平台

## 支持的Go版本
支持1.11+版本使用sdk-go

## 功能概述

- 数据点保存WritePoints
- 采集数据的转换函数ConvertValue
- LogError、LogWarn、LogInfo、LogDebug等等级日志输出
- 实现Driver接口，实现启动、重启、指令操作等函数

## 例子

```go
package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sdk"
)

// TestDriver 定义测试驱动结构体
type TestDriver struct{}

type (
	// 驱动配置信息，不同的驱动生成不同的配置信息
	config []model

	// model 模型信息
	model struct {
		ID      string `json:"id"`     // 模型id，模型唯一标识
		Device  Device `json:"device"` // 模型驱动信息
		Devices []struct {
			ID     string `json:"id"`     // 资产id，资产唯一标识
			Uid    string `json:"uid"`    // 资产的唯一编号
			Device Device `json:"device"` // 资产驱动信息
		} `json:"devices"`             // 所属模型的资产配置信息
	}

	// 驱动信息
	Device struct {
		Driver string `json:"driver"` // 驱动名称
		Tags   []struct {
			ID   string `json:"id"`   // 数据点唯一标识
			Name string `json:"name"` // 数据点名称
		} `json:"tags"`               // 驱动数据点
		Commands []struct {
			ID   string `json:"id"`   // 指令唯一标识
			Name string `json:"name"` // 指令名称
		} `json:"commands"`           // 指令配置
	}
)

// Start 驱动执行，实现Driver的Start函数
func (*TestDriver) Start(dg *sdk.DG, models []byte) error {
	log.Println("start", string(models))
	ms := config{}
	err := json.Unmarshal(models, &ms)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(ms)
	go func() {
		for {
			for _, m1 := range ms {
				if m1.Devices == nil {
					continue
				}
				for _, n1 := range m1.Devices {
					if n1.Device.Tags == nil {
						continue
					}
					fields := make(map[string]interface{})
					if m1.Device.Tags != nil {
						for _, t1 := range m1.Device.Tags {
							fields[t1.ID] = rand.Intn(100)
						}
					}
					for _, t1 := range n1.Device.Tags {
						fields[t1.ID] = rand.Intn(100)
					}
					log.Println(n1.Uid, m1.ID, n1.ID, fields)
					if err := dg.WritePoints(n1.Uid, m1.ID, n1.ID, fields); err != nil {
						dg.LogError(n1.Uid, "写数据错误")
					}

				}
			}
			time.Sleep(time.Second * 10)
		}
	}()
	return nil
}

// Reload 驱动重启，实现Driver的Reload函数
func (*TestDriver) Reload(dg *sdk.DG, models []byte) error {
	log.Println("reload", string(models))

	return nil
}

// Run 执行指令，实现Driver的Run函数
func (*TestDriver) Run(dg *sdk.DG, deviceID string, cmd []byte) error {
	log.Println("run", deviceID, string(cmd))
	return nil
}

func main() {
	// 创建采集主程序
	dg := sdk.NewDG(sdk.ServiceConfig{
		Schema: ``,
		Consul: &sdk.GCConfig{
			Host: "iot.tmis.top",
			Port: 8500,
		},
		Service: &sdk.RegistryConfig{
			ID:   "test",
			Name: "driver_test",
		},
		Driver: &sdk.DriverConfig{
			ID:   "test",
			Name: "测试",
		},
		Gateway: &sdk.GCConfig{
			Host: "iot.tmis.top",
			Port: 8010,
		},
		Mqtt: &sdk.EmqttConfig{
			Host: "iot.tmis.top",
			Port: 1883,
		},
	})
	defer func() {
		// 驱动服务停止
		if err := dg.Stop(); err != nil {
			log.Println("服务停止失败", err)
		} else {
			log.Println("服务停止成功")
		}
	}()

	// 创建测试驱动并开始执行
	err := dg.Start(new(TestDriver))
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
```