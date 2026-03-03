# AirIoT SDK Go

[English](./README_en.md) | 简体中文

AirIoT SDK Go 是用于开发 AirIoT 平台扩展服务的 Go SDK，覆盖 Driver、Algorithm、DataRelay、Flow、FlowExtension、Service、Task。

## 目录

- [模块概览](#模块概览)
- [安装](#安装)
- [快速开始](#快速开始)
- [核心接口](#核心接口)
- [Driver 配置示例](#driver-配置示例)
- [Driver 配置变更说明](#driver-配置变更说明)
- [Driver License 动态库加载](#driver-license-动态库加载)
- [示例目录](#示例目录)
- [FAQ](#faq)
- [环境要求](#环境要求)
- [许可证](#许可证)

## 模块概览

| 模块 | 包路径 | 说明 |
| --- | --- | --- |
| Driver | `github.com/air-iot/sdk-go/v4/driver` | 设备接入、点位/事件/告警上报、指令执行 |
| Algorithm | `github.com/air-iot/sdk-go/v4/algorithm` | 算法服务 |
| DataRelay | `github.com/air-iot/sdk-go/v4/data_relay` | 数据中继 |
| Flow | `github.com/air-iot/sdk-go/v4/flow` | 流程节点 |
| FlowExtension | `github.com/air-iot/sdk-go/v4/flow_extension` | 可配置流程扩展 |
| Service | `github.com/air-iot/sdk-go/v4/service` | Gin HTTP 服务 |
| Task | `github.com/air-iot/sdk-go/v4/task` | Cron 定时任务 |

## 安装

```bash
go get github.com/air-iot/sdk-go/v4
```

## 快速开始

```go
package main

import "github.com/air-iot/sdk-go/v4/driver"

func main() {
	app := driver.NewApp()
	app.Start(yourDriver)
}
```

运行仓库示例：

```bash
go run ./example/driver -config ./example/driver/etc/
go run ./example/algorithm -config ./example/algorithm/etc/
go run ./example/data_relay -config ./example/data_relay/etc/
go run ./example/flow -config ./example/flow/etc/
go run ./example/flow_extension -config ./example/flow_extension/etc/
go run ./example/service
go run ./example/task
```

## 核心接口

### Driver

```go
type Driver interface {
	Schema(ctx context.Context, app App, locale string) (schema string, err error)
	Start(ctx context.Context, app App, driverConfig []byte) (err error)
	RegisterRoutes(router *gin.RouterGroup)
	Run(ctx context.Context, app App, command *entity.Command) (result interface{}, err error)
	BatchRun(ctx context.Context, app App, command *entity.BatchCommand) (result interface{}, err error)
	WriteTag(ctx context.Context, app App, command *entity.Command) (result interface{}, err error)
	Debug(ctx context.Context, app App, debugConfig []byte) (result interface{}, err error)
	HttpProxy(ctx context.Context, app App, t string, header http.Header, data []byte) (result interface{}, err error)
	ConfigUpdate(ctx context.Context, app App, data *pb.ConfigUpdateRequest) (err error)
	Stop(ctx context.Context, app App) (err error)
}
```

### Algorithm

```go
type Service interface {
	Schema(context.Context, App, string) (result string, err error)
	Start(context.Context, App) error
	Run(ctx context.Context, app App, bts []byte) (result interface{}, err error)
	Stop(context.Context, App) error
}
```

### DataRelay

```go
type DataRelay interface {
	Start(ctx context.Context, app App, config []byte) (err error)
	HttpProxy(ctx context.Context, app App, t string, header http.Header, data []byte) (result []byte, err error)
}
```

### Flow

```go
type Flow interface {
	Handler(ctx context.Context, app App, request *Request) (result map[string]interface{}, err error)
	Debug(ctx context.Context, app App, request *DebugRequest) (result *DebugResult, err error)
}
```

### FlowExtension

```go
type Extension interface {
	Schema(ctx context.Context, app App, locale string) (schema string, err error)
	Run(ctx context.Context, app App, input []byte) (result map[string]interface{}, err error)
}
```

说明：`flow_extension` 历史包名为 `flow_extionsion`，建议使用别名导入。

## Driver 配置示例

```yaml
serviceId: your-service-id
groupId: your-group-id
project: your-project-id

driver:
  id: go-driver-demo
  name: Go Driver Demo

driverGrpc:
  enable: true
  host: localhost
  port: 9224
  health:
    requestTime: 10s
    retry: 3
  stream:
    heartbeat: 30s
  waitTime: 5s
  timeout: 600s
  limit: 100

http:
  enable: false
  host: 0.0.0.0
  port: 8080

dataFile:
  enable: true
  path: ./data.json

license: ./license

mq:
  type: mqtt
  mqtt:
    schema: tcp
    host: localhost
    port: 1883
    username: admin
    password: public

log:
  level: 4
  format: json
```

## Driver 配置变更说明

- `driverGrpc.healthRequestTime` 已调整为 `driverGrpc.health.requestTime`。
- `driverGrpc.waitTime`、`driverGrpc.timeout` 使用时长格式（如 `5s`、`600s`）。
- `dataFile.enable=true` 时，SDK 会读取 `dataFile.path` 指向的 `data.json` 作为驱动运行配置，并监听文件变化自动重载。
- 新增 `license` 配置：用于授权目录路径（注意是目录，不是单个文件）。

## Driver License 动态库加载

SDK 会按平台加载 `license_core` 动态库，并调用库函数校验驱动授权。

库名规则：

- Windows amd64: `license_core_windows_amd64.dll`
- Windows arm64: `license_core_windows_arm64.dll`
- Linux amd64: `license_core_linux_amd64.so`
- Linux arm64: `license_core_linux_arm64.so`
- Linux loong64: `license_core_linux_loong64.so`
- macOS amd64: `license_core_darwin_amd64.dylib`
- macOS arm64: `license_core_darwin_arm64.dylib`

查找顺序：

- 当前工作目录
- `./lib/`
- `./license/lib/`
- 可执行文件同目录
- 可执行文件目录下的 `lib/`

授权兜底行为：

- 当 `license` 为空或路径无效时，会按“无授权模式”校验。
- 无授权模式下允许的总 tag 上限为 `20`，超过即启动失败。
- 以下驱动 ID 免授权：`test`、`modbus`、`modbus_rtu`、`db-driver`、`driver-http-client`、`driver-mqtt-client`、`opcda`、`modbus_rtutcp`。

## 示例目录

- [example/driver](./example/driver)
- [example/driver_lazy](./example/driver_lazy)
- [example/algorithm](./example/algorithm)
- [example/data_relay](./example/data_relay)
- [example/flow](./example/flow)
- [example/flow_extension](./example/flow_extension)
- [example/service](./example/service)
- [example/task](./example/task)

## FAQ

### 为什么 `healthRequestTime` 不生效？

请改用嵌套字段：

- `driverGrpc.health.requestTime`
- `driverGrpc.health.retry`

### `license` 应该填什么？

填写授权目录路径（目录内由授权库读取所需文件）。如果留空或路径无效，会进入无授权兜底校验。

### `dataFile` 有什么作用？

用于本地驱动配置加载与热更新。启用后会读取并监听 `data.json`。

## 环境要求

- Go `>= 1.23`

## 许可证

请参考仓库许可证文件。
