# AirIoT SDK Go

English | [Chinese](./README.md)

AirIoT SDK Go is the Go SDK for building AirIoT extension services, including Driver, Algorithm, DataRelay, Flow, FlowExtension, Service, and Task modules.

## Table of Contents

- [Module Overview](#module-overview)
- [Install](#install)
- [Quick Start](#quick-start)
- [Core Interfaces](#core-interfaces)
- [Driver Configuration Example](#driver-configuration-example)
- [Driver Config Changes](#driver-config-changes)
- [Driver License Dynamic Library Loading](#driver-license-dynamic-library-loading)
- [Example Projects](#example-projects)
- [FAQ](#faq)
- [Requirements](#requirements)
- [License](#license)

## Module Overview

| Module | Package | Description |
| --- | --- | --- |
| Driver | `github.com/air-iot/sdk-go/v4/driver` | Device access, point/event/warning reporting, command handling |
| Algorithm | `github.com/air-iot/sdk-go/v4/algorithm` | Algorithm service |
| DataRelay | `github.com/air-iot/sdk-go/v4/data_relay` | Data relay and proxy |
| Flow | `github.com/air-iot/sdk-go/v4/flow` | Flow node logic |
| FlowExtension | `github.com/air-iot/sdk-go/v4/flow_extension` | Configurable flow extension |
| Service | `github.com/air-iot/sdk-go/v4/service` | Gin HTTP service |
| Task | `github.com/air-iot/sdk-go/v4/task` | Cron task runner |

## Install

```bash
go get github.com/air-iot/sdk-go/v4
```

## Quick Start

```go
package main

import "github.com/air-iot/sdk-go/v4/driver"

func main() {
	app := driver.NewApp()
	app.Start(yourDriver)
}
```

Run built-in examples:

```bash
go run ./example/driver -config ./example/driver/etc/
go run ./example/algorithm -config ./example/algorithm/etc/
go run ./example/data_relay -config ./example/data_relay/etc/
go run ./example/flow -config ./example/flow/etc/
go run ./example/flow_extension -config ./example/flow_extension/etc/
go run ./example/service
go run ./example/task
```

## Core Interfaces

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

Note: historical package name is `flow_extionsion`, so alias import is recommended.

## Driver Configuration Example

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

## Driver Config Changes

- `driverGrpc.healthRequestTime` has been replaced by `driverGrpc.health.requestTime`.
- `driverGrpc.waitTime` and `driverGrpc.timeout` use duration format (for example `5s`, `600s`).
- When `dataFile.enable=true`, SDK reads driver runtime config from `dataFile.path` (`data.json`) and hot-reloads it on file change.
- Added `license` config: this must be a directory path (not a single file path).

## Driver License Dynamic Library Loading

The SDK loads platform-specific `license_core` libraries and verifies driver license through library symbols.

Library names:

- Windows amd64: `license_core_windows_amd64.dll`
- Windows arm64: `license_core_windows_arm64.dll`
- Linux amd64: `license_core_linux_amd64.so`
- Linux arm64: `license_core_linux_arm64.so`
- Linux loong64: `license_core_linux_loong64.so`
- macOS amd64: `license_core_darwin_amd64.dylib`
- macOS arm64: `license_core_darwin_arm64.dylib`

Lookup order:

- current working directory
- `./lib/`
- `./license/lib/`
- executable directory
- `lib/` under executable directory

Fallback behavior:

- If `license` is empty or invalid, SDK switches to "no-license fallback" validation.
- In fallback mode, max total tag count is `20`; startup fails if exceeded.
- These driver IDs are exempt from license verification: `test`, `modbus`, `modbus_rtu`, `db-driver`, `driver-http-client`, `driver-mqtt-client`, `opcda`, `modbus_rtutcp`.

## Example Projects

- [example/driver](./example/driver)
- [example/driver_lazy](./example/driver_lazy)
- [example/algorithm](./example/algorithm)
- [example/data_relay](./example/data_relay)
- [example/flow](./example/flow)
- [example/flow_extension](./example/flow_extension)
- [example/service](./example/service)
- [example/task](./example/task)

## FAQ

### Why is `healthRequestTime` not working?

Use nested keys:

- `driverGrpc.health.requestTime`
- `driverGrpc.health.retry`

### What should `license` point to?

A license directory path (the native library reads files from that directory). If empty/invalid, no-license fallback is used.

### What is `dataFile` used for?

Local driver runtime config loading and hot reload of `data.json`.

## Requirements

- Go `>= 1.23`

## License

See repository license file.
