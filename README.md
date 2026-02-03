# AirIoT SDK Go

[English](./README_en.md) | 简体中文

AirIoT SDK Go 是用于开发 AirIoT 平台扩展服务的 Go 语言 SDK，提供设备驱动、算法服务、数据中继、流程插件等多种类型的扩展能力。

## 目录

- [功能特性](#功能特性)
- [快速开始](#快速开始)
- [模块类型](#模块类型)
- [配置说明](#配置说明)
- [示例代码](#示例代码)
- [许可证](#许可证)

## 功能特性

- **设备驱动 (Driver)**: 支持多种协议设备的数据采集与控制
- **算法服务 (Algorithm)**: 提供自定义算法执行能力
- **数据中继 (DataRelay)**: 实现数据转发与处理
- **流程插件 (Flow)**: 支持流程节点自定义逻辑
- **流程扩展 (FlowExtension)**: 提供流程节点 Schema 配置能力

## 快速开始

### 安装

```bash
go get github.com/air-iot/sdk-go/v4
```

### 基础使用

```go
import "github.com/air-iot/sdk-go/v4/driver"

func main() {
    // 创建应用实例
    app := driver.NewApp()
    // 启动驱动
    app.Start(yourDriver)
}
```

## 模块类型

### 1. 设备驱动 (Driver)

用于开发设备驱动程序，实现设备数据采集与控制功能。

#### Driver 接口

| 方法 | 说明 | 参数 |
|------|------|------|
| `Schema` | 返回驱动配置的 JSON Schema 定义，用于动态生成驱动配置界面 | `ctx context.Context`<br>`app App`<br>`locale string` - 国际化语言代码（如 "zh"、"en"） |
| `Start` | 启动驱动，初始化设备连接、启动数据采集等 | `ctx context.Context`<br>`app App`<br>`driverConfig []byte` - 驱动配置数据（JSON 格式），包含实例、模型及设备信息 |
| `RegisterRoutes` | 注册自定义 HTTP 路由，在 HTTP 服务启动前调用 | `router *gin.Engine` - Gin 路由引擎实例 |
| `Run` | 执行单设备指令，向设备下发控制指令或配置命令 | `ctx context.Context`<br>`app App`<br>`command *entity.Command` - 指令参数 |
| `BatchRun` | 批量执行多设备指令，一次性向多个设备下发相同的指令 | `ctx context.Context`<br>`app App`<br>`command *entity.BatchCommand` - 批量指令参数 |
| `WriteTag` | 写入数据点，向设备写入指定的数据点值 | `ctx context.Context`<br>`app App`<br>`command *entity.Command` - 写入参数 |
| `Debug` | 调试驱动，返回设备连接状态、测试数据等 | `ctx context.Context`<br>`app App`<br>`debugConfig []byte` - 调试参数（JSON 格式） |
| `HttpProxy` | HTTP 代理接口，处理驱动自定义的 HTTP 请求 | `ctx context.Context`<br>`app App`<br>`t string` - 请求接口标识<br>`header http.Header` - HTTP 请求头<br>`data []byte` - 请求体数据 |
| `ConfigUpdate` | 配置更新回调，当驱动配置在平台侧被修改时触发 | `ctx context.Context`<br>`app App`<br>`data *pb.ConfigUpdateRequest` - 配置更新请求 |
| `Stop` | 停止驱动，清理资源、关闭连接等 | `ctx context.Context`<br>`app App` |

#### Command 参数结构

```go
// Command 单设备指令参数
type Command struct {
    Table    string `json:"table"`    // 表标识
    Id       string `json:"id"`       // 设备编号
    SerialNo string `json:"serialNo"` // 流水号
    Command  []byte `json:"command"`  // 指令内容
}

// BatchCommand 批量设备指令参数
type BatchCommand struct {
    Table    string   `json:"table"`    // 表标识
    Ids      []string `json:"ids"`      // 设备编号列表
    SerialNo string   `json:"serialNo"` // 流水号
    Command  []byte   `json:"command"`  // 指令内容
}

// 指令状态常量
const (
    COMMAND_STATUS_TIMEOUT CommandStatus = "timeout" // 超时
    COMMAND_STATUS_READY   CommandStatus = "ready"   // 待处理
    COMMAND_STATUS_REVOKE  CommandStatus = "revoke"  // 撤回
    COMMAND_STATUS_SUCCESS CommandStatus = "success" // 成功
    COMMAND_STATUS_FAIL    CommandStatus = "fail"    // 失败
)
```

#### App 接口

SDK 提供的 App 接口，用于驱动与平台交互：

| 方法 | 说明 | 返回值 |
|------|------|--------|
| `GetProjectId()` | 获取项目 ID | `string` |
| `GetGroupID()` | 获取分组 ID | `string` |
| `GetServiceId()` | 获取服务 ID | `string` |
| `GetMQ()` | 获取消息队列实例 | `mq.MQ` |
| `GetRouter()` | 获取 HTTP 路由组 | `*gin.RouterGroup` |
| `StartHTTPServer()` | 启动 HTTP 服务器 | `error` |
| `WritePoints()` | 写入测点数据 | `error` |
| `SavePoints()` | 保存测点数据到指定表 | `error` |
| `WriteEvent()` | 写入事件 | `error` |
| `WriteWarning()` | 写入告警 | `error` |
| `WriteWarningRecovery()` | 写入告警恢复 | `error` |
| `FindDevice()` | 查询设备信息 | `error` |
| `RunLog()` | 写入运行日志 | `error` |
| `UpdateTableData()` | 更新表数据 | `error` |
| `LogDebug()` | 记录调试日志 | - |
| `LogInfo()` | 记录信息日志 | - |
| `LogWarn()` | 记录警告日志 | - |
| `LogError()` | 记录错误日志 | - |
| `GetCommands()` | 获取待执行指令 | `error` |
| `UpdateCommand()` | 更新指令状态 | `error` |
| `BroadcastRealtimeData()` | 广播实时数据 | `error` |

#### 数据类型

```go
// Point 存储数据点
type Point struct {
    Table      string            `json:"table"`      // 表 ID
    ID         string            `json:"id"`         // 设备编号
    CID        string            `json:"cid"`        // 子设备编号
    Fields     []Field           `json:"fields"`     // 数据点列表
    UnixTime   int64             `json:"time"`       // 数据采集时间（毫秒时间戳）
    FieldTypes map[string]string `json:"fieldTypes"` // 数据点类型映射
}

// Field 单个数据点字段
type Field struct {
    Tag   Tag         `json:"tag"`   // 数据点配置
    Value interface{} `json:"value"` // 数据采集值
}

// Event 事件数据
type Event struct {
    Table    string                 `json:"table"`    // 表 ID
    ID       string                 `json:"id"`       // 设备编号
    Type     string                 `json:"type"`     // 事件类型
    Level    string                 `json:"level"`    // 事件级别
    Content  string                 `json:"content"`  // 事件内容
    UnixTime int64                  `json:"time"`     // 事件时间（毫秒时间戳）
    Data     map[string]interface{} `json:"data"`     // 附加数据
}

// Warn 告警数据
type Warn struct {
    Table      string                 `json:"table"`      // 表 ID
    ID         string                 `json:"id"`         // 设备编号
    Type       string                 `json:"type"`       // 告警类型
    Level      string                 `json:"level"`      // 告警级别
    Content    string                 `json:"content"`    // 告警内容
    UnixTime   int64                  `json:"time"`       // 告警时间（毫秒时间戳）
    Data       map[string]interface{} `json:"data"`       // 附加数据
    RecoverTag string                 `json:"recoverTag"` // 恢复标签
}
```

### 2. 算法服务 (Algorithm)

用于开发自定义算法服务，提供算法执行能力。

#### Service 接口

| 方法 | 说明 | 参数 |
|------|------|------|
| `Schema` | 查询算法配置的 JSON Schema，支持多语言 | `ctx context.Context`<br>`app App`<br>`locale string` - 国际化语言代码 |
| `Start` | 启动算法服务，初始化算法资源 | `ctx context.Context`<br>`app App` |
| `Run` | 执行算法，处理输入数据并返回结果 | `ctx context.Context`<br>`app App`<br>`input []byte` - 执行参数，格式：`{"function":"算法名","input":{}}` |
| `Stop` | 停止算法服务，释放资源 | `ctx context.Context`<br>`app App` |

#### 输入输出格式

```go
// 算法执行输入格式
{
    "function": "algorithmName",  // 算法名称
    "input": {                    // 算法输入参数，应与 Schema 定义的格式相同
        "param1": "value1",
        "param2": "value2"
    }
}

// 算法执行输出格式
{
    "result": "value",            // 算法计算结果
    "status": "success"           // 执行状态
}
```

### 3. 数据中继 (DataRelay)

用于实现数据转发与处理功能，可作为数据桥接服务连接不同的数据源和目标。

#### DataRelay 接口

| 方法 | 说明 | 参数 |
|------|------|------|
| `Start` | 启动数据中继服务 | `ctx context.Context`<br>`app App`<br>`config []byte` - 配置数据（JSON 格式） |
| `HttpProxy` | HTTP 代理接口，处理自定义 HTTP 请求 | `ctx context.Context`<br>`app App`<br>`t string` - 请求接口标识<br>`header http.Header` - HTTP 请求头<br>`data []byte` - 请求数据 |

#### 使用场景

- 数据格式转换
- 数据转发与分发
- 第三方系统对接
- 数据缓存与队列处理

### 4. 流程插件 (Flow)

用于开发流程节点自定义逻辑，可在 AirIoT 平台的流程编排中使用。

#### Flow 接口

| 方法 | 说明 | 参数 |
|------|------|------|
| `Handler` | 执行流程插件，处理流程节点逻辑 | `ctx context.Context`<br>`app App`<br>`request *Request` - 执行参数 |
| `Debug` | 调试流程插件，返回执行日志和结果 | `ctx context.Context`<br>`app App`<br>`request *DebugRequest` - 调试参数 |

#### Request 参数结构

```go
// Request 流程执行请求
type Request struct {
    ProjectId  string `json:"projectId,omitempty"`  // 项目 ID
    FlowId     string `json:"flowId,omitempty"`     // 流程 ID
    Job        string `json:"job,omitempty"`        // 流程实例 ID
    ElementId  string `json:"elementId,omitempty"`  // 节点 ID
    ElementJob string `json:"elementJob,omitempty"` // 节点实例 ID
    Config     []byte `json:"config,omitempty"`     // 节点配置
}

// DebugRequest 流程调试请求
type DebugRequest struct {
    ProjectId string `json:"projectId,omitempty"` // 项目 ID
    FlowId    string `json:"flowId,omitempty"`    // 流程 ID
    ElementId string `json:"elementId,omitempty"` // 节点 ID
    Config    []byte `json:"config,omitempty"`    // 节点配置
}

// DebugResult 调试结果
type DebugResult struct {
    Logs  []Syslog               `json:"logs"`  // 调试日志列表
    Value map[string]interface{} `json:"value"` // 调试输出值
}

// Syslog 系统日志
type Syslog struct {
    Level string `json:"level"` // 日志级别
    Time  string `json:"time"`  // 日志时间
    Msg   string `json:"msg"`   // 日志消息
}
```

### 5. 流程扩展 (FlowExtension)

提供流程节点的 Schema 配置和执行能力，用于创建可配置的流程扩展服务。

#### Extension 接口

| 方法 | 说明 | 参数 |
|------|------|------|
| `Schema` | 查询扩展服务的 JSON Schema 定义，支持多语言 | `ctx context.Context`<br>`app App`<br>`locale string` - 国际化语言代码 |
| `Run` | 执行扩展服务，处理输入数据并返回结果 | `ctx context.Context`<br>`app App`<br>`input []byte` - 执行参数，应与 Schema 格式相同 |

#### 与 Flow 的区别

- **Flow**: 直接在流程编排中使用，通过代码实现节点逻辑
- **FlowExtension**: 提供可配置的扩展服务，通过 Schema 定义配置界面，配置数据作为输入执行

## 使用示例

### Driver 完整示例

```go
package main

import (
    "context"
    "github.com/air-iot/sdk-go/v4/driver"
    "github.com/air-iot/sdk-go/v4/driver/entity"
)

type MyDriver struct{}

// Schema 返回驱动配置的 Schema 定义
func (d *MyDriver) Schema(ctx context.Context, app driver.App, locale string) (string, error) {
    return `{
        "type": "object",
        "properties": {
            "host": {
                "type": "string",
                "title": "主机地址",
                "default": "localhost"
            },
            "port": {
                "type": "integer",
                "title": "端口号",
                "default": 502
            }
        }
    }`, nil
}

// Start 启动驱动
func (d *MyDriver) Start(ctx context.Context, app driver.App, driverConfig []byte) error {
    // 解析配置
    // 初始化设备连接
    // 启动数据采集
    return nil
}

// Run 执行单设备指令
func (d *MyDriver) Run(ctx context.Context, app driver.Driver, command *entity.Command) (interface{}, error) {
    // 解析指令内容
    // 向设备下发指令
    // 返回执行结果
    return map[string]interface{}{"status": "success"}, nil
}

// BatchRun 批量执行多设备指令
func (d *MyDriver) BatchRun(ctx context.Context, app driver.Driver, command *entity.BatchCommand) (interface{}, error) {
    // 批量处理多个设备的指令
    return map[string]interface{}{"status": "success"}, nil
}

// WriteTag 写入数据点
func (d *MyDriver) WriteTag(ctx context.Context, app driver.Driver, command *entity.Command) (interface{}, error) {
    // 向设备写入数据点
    return map[string]interface{}{"status": "success"}, nil
}

// Debug 调试驱动
func (d *MyDriver) Debug(ctx context.Context, app driver.Driver, debugConfig []byte) (interface{}, error) {
    // 返回调试信息
    return map[string]interface{}{
        "connections": []string{"device1", "device2"},
        "status": "online",
    }, nil
}

// HttpProxy HTTP 代理接口
func (d *MyDriver) HttpProxy(ctx context.Context, app driver.Driver, t string, header http.Header, data []byte) (interface{}, error) {
    // 处理自定义 HTTP 请求
    return map[string]interface{}{"result": "ok"}, nil
}

// ConfigUpdate 配置更新回调
func (d *MyDriver) ConfigUpdate(ctx context.Context, app driver.Driver, data *pb.ConfigUpdateRequest) error {
    // 处理配置更新
    return nil
}

// Stop 停止驱动
func (d *MyDriver) Stop(ctx context.Context, app driver.Driver) error {
    // 清理资源、关闭连接
    return nil
}

// RegisterRoutes 注册自定义 HTTP 路由
func (d *MyDriver) RegisterRoutes(router *gin.Engine) {
    router.GET("/custom", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "custom route"})
    })
}

func main() {
    app := driver.NewApp()
    app.Start(&MyDriver{})
}
```

### Algorithm 完整示例

```go
package main

import (
    "context"
    "github.com/air-iot/sdk-go/v4/algorithm"
)

type MyAlgorithm struct{}

// Schema 返回算法配置的 Schema 定义
func (a *MyAlgorithm) Schema(ctx context.Context, app algorithm.App, locale string) (string, error) {
    return `{
        "type": "object",
        "properties": {
            "threshold": {
                "type": "number",
                "title": "阈值",
                "default": 100
            }
        }
    }`, nil
}

// Start 启动算法服务
func (a *MyAlgorithm) Start(ctx context.Context, app algorithm.App) error {
    // 初始化算法资源
    return nil
}

// Run 执行算法
func (a *MyAlgorithm) Run(ctx context.Context, app algorithm.App, input []byte) (interface{}, error) {
    // 解析输入数据
    // 执行算法逻辑
    // 返回计算结果
    return map[string]interface{}{
        "result": 42,
        "status": "success",
    }, nil
}

// Stop 停止算法服务
func (a *MyAlgorithm) Stop(ctx context.Context, app algorithm.App) error {
    // 释放资源
    return nil
}

func main() {
    app := algorithm.NewApp()
    app.Start(&MyAlgorithm{})
}
```

## 配置说明

### Driver 配置示例

```yaml
serviceId: your-service-id
project: your-project
driver:
  id: go-driver-demo
  name: 演示驱动
log:
  level: 5
  format: json
driverGrpc:
  enable: true
  host: localhost
  port: 9224
  healthRequestTime: 10s
  waitTime: 5s
http:
  enable: true
  host: 0.0.0.0
  port: 8080
dataFile:
  enable: true
  path: ./data.json
```

### Algorithm 配置示例

```yaml
serviceId: your-service-id
algorithm:
  id: go-algorithm-demo
  name: 演示算法
log:
  level: 5
  format: json
algorithmGrpc:
  host: localhost
  port: 9236
  healthRequestTime: 10
  waitTime: 5
api:
  gateway: http://localhost:3030/rest
  gatewayGrpc: localhost:9224
```

### 配置参数说明

#### Driver 配置参数

| 参数 | 说明 | 类型 | 默认值 |
|------|------|------|--------|
| `serviceId` | 服务 ID，唯一标识服务实例 | string | 自动生成 |
| `project` | 项目 ID | string | - |
| `driver.id` | 驱动标识 | string | - |
| `driver.name` | 驱动名称 | string | - |
| `driverGrpc.enable` | 是否启用 gRPC 连接 | bool | false |
| `driverGrpc.host` | gRPC 服务器地址 | string | - |
| `driverGrpc.port` | gRPC 服务器端口 | int | 9224 |
| `driverGrpc.healthRequestTime` | 健康检查请求间隔 | duration | 10s |
| `driverGrpc.waitTime` | 重连等待时间 | duration | 5s |
| `log.level` | 日志级别 (1-5) | int | 4 |
| `log.format` | 日志格式 (json/text) | string | json |
| `mq.type` | 消息队列类型 (local/mqtt/kafka) | string | local |
| `http.enable` | 是否启用 HTTP 服务 | bool | false |
| `http.host` | HTTP 服务监听地址 | string | 0.0.0.0 |
| `http.port` | HTTP 服务端口 | int | 8080 |
| `dataFile.enable` | 是否启用本地数据文件 | bool | false |
| `dataFile.path` | data.json 文件路径 | string | ./data.json |
| `pprof.enable` | 是否启用 pprof 性能分析 | bool | false |
| `pprof.host` | pprof 监听地址 | string | - |
| `pprof.port` | pprof 监听端口 | int | - |

#### Algorithm 配置参数

| 参数 | 说明 | 类型 | 默认值 |
|------|------|------|--------|
| `serviceId` | 服务 ID | string | 自动生成 |
| `algorithm.id` | 算法标识 | string | - |
| `algorithm.name` | 算法名称 | string | - |
| `algorithmGrpc.host` | gRPC 服务器地址 | string | - |
| `algorithmGrpc.port` | gRPC 服务器端口 | int | 9236 |
| `algorithmGrpc.limit` | 并发限制 | int | 100 |
| `algorithm.timeout` | 执行超时时间（秒） | int | 600 |
| `api.gateway` | API 网关地址 | string | - |
| `api.gatewayGrpc` | API gRPC 地址 | string | - |
| `api.type` | API 类型 | string | project |
| `api.projectId` | 项目 ID | string | default |
| `api.ak` | Access Key | string | - |
| `api.sk` | Secret Key | string | - |

#### 日志级别说明

| 级别 | 值 | 说明 |
|------|-----|------|
| Trace | 1 | 最详细的日志信息 |
| Debug | 2 | 调试信息 |
| Info | 3 | 一般信息 |
| Warn | 4 | 警告信息 |
| Error | 5 | 错误信息 |

#### 消息队列配置

##### MQTT 配置

```yaml
mq:
  type: mqtt
  mqtt:
    host: localhost
    port: 1883
    username: user
    password: pass
    clientID: driver-client
```

##### Kafka 配置

```yaml
mq:
  type: kafka
  kafka:
    brokers:
      - localhost:9092
    topic: airiot-driver
```

## 示例代码

完整的示例代码请参考 [example](./example) 目录：

- [driver 示例](./example/driver) - 设备驱动完整示例
- [algorithm 示例](./example/algorithm) - 算法服务完整示例
- [data_relay 示例](./example/data_relay) - 数据中继完整示例
- [flow 示例](./example/flow) - 流程插件完整示例
- [flow_extension 示例](./example/flow_extension) - 流程扩展完整示例

## 最佳实践

### Driver 开发建议

1. **连接管理**: 使用连接池管理设备连接，避免频繁创建和销毁连接
2. **并发控制**: 对设备访问进行并发控制，避免同时操作同一设备
3. **错误处理**: 实现完善的错误处理机制，记录详细的错误日志
4. **资源清理**: 在 `Stop` 方法中正确释放所有资源
5. **配置验证**: 在 `Start` 方法中验证配置参数的有效性

### 数据上报建议

1. **批量上报**: 对于高频采集的数据，建议批量上报以减少网络开销
2. **时间戳**: 使用设备实际采集时间，而非上报时间
3. **数据类型**: 正确设置数据类型，确保数据精度
4. **异常处理**: 对采集异常的数据进行标记和处理

### 性能优化

1. **异步处理**: 使用 goroutine 处理耗时操作
2. **缓存策略**: 合理使用缓存减少重复计算
3. **连接复用**: 复用网络连接，减少握手开销
4. **资源限制**: 设置合理的并发限制，避免资源耗尽

## 常见问题

### Q: 如何调试驱动？

A: 可以通过以下方式调试驱动：
1. 使用 `Debug` 方法返回调试信息
2. 启用 HTTP 服务，添加自定义调试接口
3. 使用 `LogDebug`/`LogInfo` 等方法记录日志
4. 启用 pprof 进行性能分析

### Q: 驱动如何处理配置更新？

A: 实现 `ConfigUpdate` 方法，当配置在平台侧被修改时会自动调用此方法。

### Q: 如何实现设备断线重连？

A: 在驱动中实现心跳机制和重连逻辑：
1. 定期检查设备连接状态
2. 检测到断线时自动重连
3. 使用 `WriteWarning` 上报设备离线告警

### Q: 数据上报失败怎么办？

A: 检查以下几点：
1. 确认 MQ 配置正确
2. 检查网络连接
3. 查看错误日志定位问题
4. 考虑实现本地缓存，网络恢复后重新上报

### Q: 如何处理指令超时？

A:
1. 设置合理的指令执行超时时间
2. 使用 context 控制 goroutine 生命周期
3. 超时后更新指令状态为 `timeout`

## 故障排除

### 启动失败

1. 检查配置文件格式是否正确
2. 确认 Go 版本满足要求（>= 1.23）
3. 查看日志中的错误信息
4. 验证依赖包是否完整安装

### 连接失败

1. 确认网络连接正常
2. 检查防火墙设置
3. 验证目标地址和端口配置
4. 检查认证信息是否正确

### 性能问题

1. 启用 pprof 分析性能瓶颈
2. 检查 goroutine 泄漏
3. 优化数据批量上报策略
4. 调整并发限制参数

## 技术支持

如有问题或建议，请通过以下方式联系：

- 提交 Issue
- 查看项目文档
- 联系技术支持团队

## 依赖要求

- Go 1.23 或更高版本

## 更新日志

### v4
- 重构驱动接口，增加更多功能
- 优化性能和稳定性
- 改进错误处理机制
- 增加更多配置选项

## 许可证

请参阅项目许可证文件获取详细信息。
