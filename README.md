# sdk

支持用户通过go开发数据采集驱动，并接入云组态监控平台

## 功能概述

- 数据保存函数WritePoints
- 采集数据的转换函数ConvertValue
- 日志函数LogError、LogWarn、LogInfo、LogDebug
- 实现Driver接口，实现Start、Reload、Run等函数。Start根据配置实现采集驱动的业务逻辑，Reload重新加载配置后实现采集驱动的业务逻辑，Run实现控制指令操作

## 安装

```
go get -v github.com/casic-iot/sdk-go
```

要求 Go >= 1.11

## 使用

```go
// 这个例子展示如何使用go开发工具包开发驱动
package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/casic-iot/sdk-go"
)

// TestDriver 定义测试驱动结构体
type TestDriver struct{
	ctx    context.Context
	cancel context.CancelFunc
}

// Start 驱动执行，实现Driver的Start函数
func (*TestDriver) Start(dg *sdk.DG, models []byte) error {
	log.Println("开始", string(models))
	return nil
}

// Reload 驱动重启，实现Driver的Reload函数
func (*TestDriver) Reload(dg *sdk.DG, models []byte) error {
	log.Println("重启", string(models))
	return nil
}

// Run 执行指令，实现Driver的Run函数
func (*TestDriver) Run(dg *sdk.DG, deviceID string, cmd []byte) error {
	log.Println("指令", deviceID, string(cmd))
	return nil
}

func main() {
	// 创建采集主程序
	dg := sdk.NewDG(sdk.ServiceConfig{
		Schema: ``,
		Consul: &sdk.GCConfig{
			Host: "localhost",
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
			Host: "localhost",
			Port: 8010,
		},
		Mqtt: &sdk.EmqttConfig{
			Host: "localhost",
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
		log.Printf("接收信号 %s\n", sig)
	}
}
```

## 消息工具

### 基本介绍
MQTTBox 是一个带有可视化的界面的 MQTT 的客户端工具，它具有如下特点：

- 支持 TCP、TLS、Web Sockets 和安全的 Web Sockets 连接 MQTT 服务器
- 支持各种 MQTT 客户端的设置
- 支持发布和订阅多个主题
- 支持主题的单级和多级订阅
- 复制/重新发布有效负载
- 支持查看每个主题已发布/已订阅消息的历史记录

### 下载安装
软件支持在 Windows、Mac 和 Linux 上面运行，我们到其官网选择合适的版本下载安装即可，下载 [MQTTBox](http://workswithweb.com/html/mqttbox/downloads.html)

### 消息订阅
#### 数据

- data/#

#### 日志

- Debug: logs/debug/资产唯一编号
- Info: logs/info/资产唯一编号
- Warn: logs/warn/资产唯一编号
- Error: logs/error/资产唯一编号
