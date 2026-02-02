# AIRIOT Go SDK 开发文档

## 项目概述

**项目名称**: sdk-go
**版本**: v4
**官网**: https://airiot.cn
**文档**: https://docs.airiot.cn
**Go开发文档**: https://docs.airiot.cn/category/go

AIRIOT Go SDK 是云平台 Go 语言开发驱动包，用于开发设备驱动、算法、流程插件、数据中继等服务。

### 系统要求

- Go 1.23.0 或更高版本
- 支持 Linux、macOS、Windows、LoongArch 等平台

## 目录结构

```
sdk-go/
├── algorithm/           # 算法服务模块
├── conn/               # 连接层组件
│   ├── mq/            # 消息队列抽象层
│   │   ├── mqtt/      # MQTT 客户端
│   │   ├── rabbit/    # RabbitMQ 客户端
│   │   ├── kafka/     # Kafka 客户端
│   │   └── local.go   # 本地内存队列
│   ├── sql/           # SQL 数据库连接
│   ├── tcp/           # TCP 连接
│   └── websocket/     # WebSocket 连接
├── data_relay/        # 数据中继服务模块
├── driver/            # 驱动服务模块（核心）
│   ├── convert/       # 数据转换工具
│   ├── entity/        # 实体定义
│   ├── grpc/          # gRPC 客户端
│   └── html/          # Web 配置界面（Next.js + shadcn/ui）
├── etcd/              # ETCD 配置中心
├── example/           # 示例代码
│   ├── algorithm/     # 算法示例
│   ├── data_relay/    # 数据中继示例
│   ├── driver/        # 驱动示例
│   ├── driver_lazy/   # 懒加载驱动示例
│   ├── flow/          # 流程插件示例
│   ├── flow_extension/# 流程扩展示例
│   ├── service/       # 服务示例
│   └── task/          # 任务示例
├── flow/              # 流程插件模块
├── flow_extension/    # 流程扩展模块
├── service/           # 通用服务模块
├── task/              # 定时任务模块
└── utils/             # 工具类
    ├── cipherx/       # 加密工具
    ├── decrypt/       # 解密工具
    ├── numberx/       # 数值转换
    └── serial/        # 串口工具
```

## 核心模块

### 1. Driver（驱动模块）- 核心模块

#### Driver 接口

```go
type Driver interface {
    // Schema 查询返回驱动配置schema内容
    Schema(ctx context.Context, app App, locale string) (schema string, err error)

    // Start 驱动启动
    Start(ctx context.Context, app App, driverConfig []byte) (err error)

    // Run 运行指令,向设备写入数据
    Run(ctx context.Context, app App, command *entity.Command) (result interface{}, err error)

    // BatchRun 批量运行指令,向多设备写入数据
    BatchRun(ctx context.Context, app App, command *entity.BatchCommand) (result interface{}, err error)

    // WriteTag 数据点写入
    WriteTag(ctx context.Context, app App, command *entity.Command) (result interface{}, err error)

    // Debug 调试驱动
    Debug(ctx context.Context, app App, debugConfig []byte) (result interface{}, err error)

    // HttpProxy 代理接口
    HttpProxy(ctx context.Context, app App, t string, header http.Header, data []byte) (result interface{}, err error)

    // ConfigUpdate 配置更新
    ConfigUpdate(ctx context.Context, app App, data *pb.ConfigUpdateRequest) (err error)

    // Stop 驱动停止处理
    Stop(ctx context.Context, app App) (err error)
}
```

#### App 接口方法

```go
type App interface {
    // 启动驱动
    Start(Driver)

    // 获取项目/服务ID
    GetProjectId() string
    GetGroupID() string
    GetServiceId() string
    GetMQ() mq.MQ

    // 写入数据点
    WritePoints(context.Context, entity.Point) error
    SavePoints(ctx context.Context, tableId string, data *entity.WritePoint) error

    // 写入事件和报警
    WriteEvent(context.Context, entity.Event) error
    WriteWarning(context.Context, entity.Warn) error
    WriteWarningRecovery(ctx context.Context, tableId, dataId string, w entity.WarnRecovery) error

    // 查询和更新
    FindDevice(ctx context.Context, table, id string, ret interface{}) error
    RunLog(context.Context, entity.Log) error
    UpdateTableData(ctx context.Context, table, id string, custom map[string]interface{}) error

    // 指令管理
    GetCommands(ctx context.Context, table, id string, ret interface{}) error
    UpdateCommand(ctx context.Context, id string, data entity.DriverInstruct) error

    // 日志记录
    LogDebug(table, id string, msg interface{})
    LogInfo(table, id string, msg interface{})
    LogWarn(table, id string, msg interface{})
    LogError(table, id string, msg interface{})
}
```

### 2. Algorithm（算法模块）

```go
type Service interface {
    // Schema 查询算法配置schema
    Schema(context.Context, App, string) (result string, err error)

    // Start 启动算法服务
    Start(context.Context, App) error

    // Run 执行算法服务
    Run(ctx context.Context, app App, bts []byte) (result interface{}, err error)

    // Stop 停止算法服务
    Stop(context.Context, App) error
}
```

### 3. Flow（流程插件模块）

```go
type Flow interface {
    // Handler 执行流程插件
    Handler(ctx context.Context, app App, request *Request) (result map[string]interface{}, err error)

    // Debug 调试流程插件
    Debug(ctx context.Context, app App, request *DebugRequest) (result *DebugResult, err error)
}
```

### 4. Flow Extension（流程扩展模块）

```go
type Extension interface {
    // Schema 查询schema
    Schema(ctx context.Context, app App, locale string) (schema string, err error)

    // Run 执行算法服务
    Run(ctx context.Context, app App, input []byte) (result map[string]interface{}, err error)
}
```

### 5. Data Relay（数据中继模块）

```go
type DataRelay interface {
    // Start 启动服务
    Start(ctx context.Context, app App, config []byte) (err error)

    // HttpProxy 代理接口
    HttpProxy(ctx context.Context, app App, t string, header http.Header, data []byte) (result []byte, err error)
}
```

### 6. Task（定时任务模块）

```go
type Task interface {
    Start(App) error
    Stop(App) error
}
```

### 7. Service（通用服务模块）

```go
type Service interface {
    Start(App) error
    Stop(App) error
}
```

## 实体类型定义

### 数据点相关

#### Point - 存储数据点

```go
type Point struct {
    Table      string            // 表id
    ID         string            // 设备编号
    CID        string            // 子设备编号
    Fields     []Field           // 数据点
    UnixTime   int64             // 数据采集时间（毫秒）
    FieldTypes map[string]string // 数据点类型
}
```

#### WritePoint - 写入数据点

```go
type WritePoint struct {
    ID         string                 // 设备ID
    CID        string                 // 子设备编号
    Source     string                 // 标识源类型
    Fields     map[string]interface{} // 数据点
    UnixTime   int64                  // 时间（毫秒）
    FieldTypes map[string]string      // 数据点类型
}
```

#### Field - 字段

```go
type Field struct {
    Tag   Tag         // 数据点配置
    Value interface{} // 数据采集值
}
```

#### Tag - 数据点配置

```go
type Tag struct {
    ID            string      // ID
    Name          string      // 自定义名称
    TagValue      *TagValue   // 值计算相关
    Fixed         *int32      // 固定小数位
    Mod           *float64    // 取模
    Range         *Range      // 范围配置
    BaseValFormat string      // 基础值格式
}
```

### 指令相关

#### Command - 单设备指令

```go
type Command struct {
    Table    string // 表标识
    Id       string // 设备编号
    SerialNo string // 流水号
    Command  []byte // 指令内容
}
```

#### BatchCommand - 批量指令

```go
type BatchCommand struct {
    Table    string   // 表标识
    Ids      []string // 设备编号列表
    SerialNo string   // 流水号
    Command  []byte   // 指令内容
}
```

#### CommandStatus - 指令状态

```go
type CommandStatus string

const (
    COMMAND_STATUS_TIMEOUT CommandStatus = "timeout" // 超时
    COMMAND_STATUS_READY   CommandStatus = "ready"   // 待处理
    COMMAND_STATUS_REVOKE  CommandStatus = "revoke"  // 撤回
    COMMAND_STATUS_SUCCESS CommandStatus = "success" // 成功
    COMMAND_STATUS_FAIL    CommandStatus = "fail"    // 失败
)
```

### 事件相关

#### Event - 事件

```go
type Event struct {
    Table    string      // 表id
    ID       string      // 设备编号
    EventID  string      // 事件ID
    UnixTime int64       // 数据采集时间（毫秒）
    Data     interface{} // 事件数据
}
```

### 报警相关

#### Warn - 报警

```go
type Warn struct {
    ID          string            // 报警ID
    TableId     string            // 表id
    TableDataId string            // 设备id
    Level       string            // 报警级别
    Ruleid      string            // 规则id
    Fields      []WarnTag         // 报警数据点
    WarningType []string          // 报警类型
    Processed   WarnProcessed     // 处理状态
    Time        *time.Time        // 报警时间
    Alert       bool              // 是否告警
    Status      WarnStatus        // 确认状态
    Handle      bool              // 是否处理
    Desc        string            // 描述
    I18nProp    map[string]string // 国际化属性
}
```

#### WarnRecovery - 报警恢复

```go
type WarnRecovery struct {
    ID   []string         // 报警ID列表
    Data WarnRecoveryData // 恢复数据
}
```

## 配置说明

### Driver 配置结构

```go
type Config struct {
    ServiceID   string       // 服务ID
    GroupID     string       // 组ID
    Project     string       // 项目ID
    Mode        Mode         // 运行模式：local（本地模式）/ normal（正常模式）
    Driver      struct {
        ID   string // 驱动ID
        Name string // 驱动名称
    }
    DriverGrpc  grpc.Config    // gRPC配置（正常模式下使用）
    HTTP        struct {
        Host   string // HTTP服务器地址
        Port   string // HTTP服务器端口
    }
    DataConfig  string       // 配置文件路径（data.json，仅在本地模式下使用）
    Log         logger.Config  // 日志配置
    MQ          mq.Config      // 消息队列配置
    Pprof       struct {
        Enable bool   // 是否启用pprof
        Host   string // 主机
        Port   string // 端口
    }
    EtcdConfig  string        // ETCD配置路径
    Etcd        etcd.Config   // ETCD配置
    API         apiConfig.Config // API配置
}
```

### Mode（运行模式）

SDK 支持两种运行模式：

| 模式 | 说明 | 特性 |
|------|------|------|
| **local** | 本地模式 | • 启动 HTTP 服务器和 Web 配置界面<br>• 处理 data.json 配置文件<br>• 支持设备状态监控和 WebSocket 推送<br>• 适用于本地开发和测试 |
| **normal** | 正常模式 | • 连接 gRPC 服务器<br>• 不处理 data.json<br>• 不启动 HTTP 服务器<br>• 适用于生产环境部署 |

**默认值**: `normal`

**使用场景：**
- **local 模式**：开发测试阶段，需要使用 Web 界面配置驱动、查看设备状态等
- **normal 模式**：生产环境，驱动连接到云平台的 gRPC 服务器

### HTTP 服务器和 Web 配置界面

v4 版本新增内置 HTTP 服务器，提供可视化 Web 配置界面（**仅在 local 模式下可用**）：

**功能特性：**
- 基于 Next.js + shadcn/ui 构建的现代化配置界面
- 支持模型、设备、点位、指令、事件的可视化配置
- 基于 Schema 驱动的动态表单生成
- 支持亮色/暗色主题切换
- 支持中英文国际化
- 配置自动保存到 data.json
- 设备状态实时监控（WebSocket 推送）

**HTTP 配置示例（local 模式）：**
```yaml
mode: local
http:
  host: ""
  port: "8080"
```

访问 `http://localhost:8080` 即可打开 Web 配置界面。

**注意**：HTTP 服务器仅在 `mode: local` 时启动，`normal` 模式下不会启动 HTTP 服务。

### 本地消息队列

SDK 支持本地内存消息队列（适用于本地测试和开发）：

```yaml
mq:
  type: local  # 本地内存队列
```

### 配置文件示例

#### Local 模式配置（本地开发测试）

```yaml
# 运行模式
mode: local

# 服务配置
serviceId: 64f847d563d1482d33753c25
project: zq

# 驱动配置
driver:
  id: go-driver-mqtt-demo
  name: 测试驱动

# HTTP服务器配置（Web配置界面）
http:
  host: ""
  port: "8080"

# 配置文件路径
dataConfig: ./data.json

# 消息队列配置（本地模式可使用 local 队列）
mq:
  type: local  # mqtt / rabbit / kafka / local

# 日志配置
log:
  level: 5      # 1:Error, 2:Warn, 3:Info, 4:Debug, 5:Trace
  format: console  # json / console
```

访问 `http://localhost:8080` 即可打开 Web 配置界面。

#### Normal 模式配置（生产环境）

```yaml
# 运行模式
mode: normal

# 服务配置
serviceId: 64f847d563d1482d33753c25
project: zq
groupId: group-001

# 驱动配置
driver:
  id: go-driver-mqtt-demo
  name: 测试驱动

# gRPC配置（连接云平台）
driverGrpc:
  host: grpc-server
  port: 9224
  healthRequestTime: 10s
  waitTime: 5s

# 消息队列配置
mq:
  type: mqtt  # mqtt / rabbit / kafka
  mqtt:
    host: mqtt-server
    port: 1883
    username: admin
    password: public

# 日志配置
log:
  level: 3      # 1:Error, 2:Warn, 3:Info, 4:Debug, 5:Trace
  format: json  # json / console

# ETCD配置
etcdConfig: /airiot/config/pro.json
etcd:
  endpoints:
    - etcd-server:2379
  dialTimeout: 60
  username: root
  password: ""
```

## 快速开始

### 1. 安装依赖

```bash
go get github.com/air-iot/sdk-go/v4
```

### 2. 创建驱动

```go
package main

import (
    "context"
    "github.com/air-iot/sdk-go/v4/driver"
    "github.com/air-iot/sdk-go/v4/driver/entity"
)

// MyDriver 实现 Driver 接口
type MyDriver struct{}

// Schema 返回驱动配置Schema
func (d *MyDriver) Schema(ctx context.Context, app driver.App, locale string) (string, error) {
    return `{
        "type": "object",
        "properties": {
            "host": {"type": "string", "title": "主机地址"},
            "port": {"type": "integer", "title": "端口"}
        }
    }`, nil
}

// Start 启动驱动
func (d *MyDriver) Start(ctx context.Context, app driver.App, driverConfig []byte) error {
    // 1. 解析驱动配置
    // 2. 建立设备连接
    // 3. 启动数据采集

    // 示例：写入数据点
    app.WritePoints(ctx, entity.Point{
        Table: "表ID",
        ID:    "设备ID",
        Fields: []entity.Field{
            {Tag: tag1, Value: 1.23},
            {Tag: tag2, Value: "value"},
        },
        UnixTime: time.Now().UnixMilli(),
    })

    return nil
}

// Run 执行指令
func (d *MyDriver) Run(ctx context.Context, app driver.App, command *entity.Command) (interface{}, error) {
    // 处理下发的指令
    return "执行成功", nil
}

// Stop 停止驱动
func (d *MyDriver) Stop(ctx context.Context, app driver.App) error {
    // 清理资源
    return nil
}

func main() {
    // 创建并启动驱动
    d := &MyDriver{}
    driver.NewApp().Start(d)
}
```

## 数据上报

### 写入数据点

```go
app.WritePoints(ctx, entity.Point{
    Table: "表ID",
    ID:    "设备ID",
    Fields: []entity.Field{
        {Tag: tag1, Value: 1.23},
        {Tag: tag2, Value: "value"},
    },
    UnixTime: time.Now().UnixMilli(),
})
```

### 写入事件

```go
app.WriteEvent(ctx, entity.Event{
    Table:   "表ID",
    ID:      "设备ID",
    EventID: "事件ID",
    Data:    eventData,
})
```

### 写入报警

```go
app.WriteWarning(ctx, entity.Warn{
    ID:          "报警ID",
    TableId:     "表ID",
    TableDataId: "设备ID",
    Level:       "报警级别",
    Fields:      []entity.WarnTag{...},
})
```

### 写入报警恢复

```go
app.WriteWarningRecovery(ctx, "表ID", "设备ID", entity.WarnRecovery{
    ID: []string{"报警ID1", "报警ID2"},
    Data: entity.WarnRecoveryData{
        Fields: []entity.WarnTag{...},
    },
})
```

## 连接层组件

### 消息队列

SDK 支持多种消息队列：

- **MQTT**: Eclipse Paho MQTT Client
- **RabbitMQ**: amqp091-go
- **Kafka**: IBM Sarama

#### MQ 接口

```go
type MQ interface {
    // 发布消息
    Publish(ctx context.Context, topicParams []string, payload []byte) error

    // 消费消息
    Consume(ctx context.Context, topicParams []string, splitN int, handler Handler) error

    // 取消订阅
    UnSubscription(ctx context.Context, sub []string) error

    // 设置回调
    Callback(Callback)
}
```

### 其他连接组件

- **TCP**: `/conn/tcp/` - TCP 连接
- **WebSocket**: `/conn/websocket/` - WebSocket 连接
- **SQL**: `/conn/sql/` - SQL 数据库连接
- **串口**: `/utils/serial/` - 串口通信（支持 LoongArch）

## 工具类

### cipherx - 加密工具

提供数据加密功能。

### decrypt - 解密配置

支持配置文件中的加密字段解密。

### numberx - 数值转换

提供数值类型转换功能。

### serial - 串口工具

提供串口通信功能，支持 LoongArch 架构。

## 主要依赖

```
github.com/air-iot/api-client-go/v4    v4.8.17    # API客户端
github.com/eclipse/paho.mqtt.golang    v1.4.3     # MQTT客户端
github.com/rabbitmq/amqp091-go         v1.9.0     # RabbitMQ客户端
github.com/IBM/sarama                  v1.42.1    # Kafka客户端
google.golang.org/grpc                 v1.75.0    # gRPC框架
github.com/gin-gonic/gin               v1.9.1     # HTTP框架
github.com/spf13/viper                 v1.20.1    # 配置管理
go.etcd.io/etcd/client/v3              v3.5.15    # ETCD客户端
github.com/dop251/goja                            # JavaScript引擎
```

## 开发指南

### 驱动开发流程

1. 实现 `Driver` 接口
2. 创建 `main.go`，调用 `driver.NewApp().Start(driver实例)`
3. 配置 `config.yaml`
4. 实现核心方法：
   - `Schema()` - 返回驱动配置Schema
   - `Start()` - 初始化连接，启动数据采集
   - `Run()` - 执行下发的指令
   - `Stop()` - 清理资源

### 数据类型常量

```go
const (
    String     = "string"
    Float      = "float"
    Integer    = "integer"
    Boolean    = "boolean"
    BooleanRaw = "boolean_raw"
)
```

### 日志级别

```go
1: Error
2: Warn
3: Info
4: Debug
5: Trace
```

## 项目特点

1. **模块化设计**: 各个服务模块独立，易于扩展
2. **多协议支持**: 支持 MQTT、RabbitMQ、Kafka、本地内存队列等消息队列
3. **gRPC通信**: 使用gRPC进行服务间通信（可选配置）
4. **内置HTTP服务器**: 提供可视化Web配置界面，支持模型、设备、点位等配置
5. **灵活配置**: 支持YAML、ENV多种配置方式，配置自动持久化
6. **完善日志**: 集成结构化日志，支持分级输出
7. **健康检查**: 内置健康检查机制
8. **流式处理**: 支持流式数据处理
9. **多平台支持**: 支持Linux、macOS、Windows、LoongArch等平台
10. **国际化支持**: Web界面支持中英文切换
11. **主题切换**: Web界面支持亮色/暗色主题

## 常见问题

### Q: Local 和 Normal 模式有什么区别？

A: 两种模式的主要区别：

| 功能 | Local 模式 | Normal 模式 |
|------|-----------|-------------|
| HTTP 服务器 | ✅ 启动 | ❌ 不启动 |
| Web 配置界面 | ✅ 可用 | ❌ 不可用 |
| data.json 处理 | ✅ 支持 | ❌ 不处理 |
| 设备状态监控 | ✅ 支持 | ❌ 不监控 |
| gRPC 连接 | ❌ 不连接 | ✅ 连接云平台 |

**建议**：
- 开发测试阶段使用 `local` 模式，便于配置和调试
- 生产环境使用 `normal` 模式，连接云平台 gRPC 服务器

### Q: 如何启用 Web 配置界面？

A: Web 配置界面仅在 `local` 模式下可用。配置方法：

A: 在 `config.yaml` 中配置 `http` 节点：
```yaml
http:
  enable: true
  host: ""
  port: "8080"
```
然后访问 `http://localhost:8080` 即可打开配置界面。

### Q: 如何使用本地消息队列？

A: 将 `mq.type` 设置为 `local`：
```yaml
mq:
  type: local
```
本地队列适用于开发测试，数据存储在内存中。

### Q: gRPC 配置是否必需？

A: 这取决于运行模式：
- **local 模式**：gRPC 配置不是必需的，驱动主要使用 HTTP 服务器提供配置界面
- **normal 模式**：gRPC 配置是必需的，驱动需要通过 gRPC 连接云平台服务器

### Q: 两种模式可以混用吗？

A: 不可以。每次运行只能选择一种模式：
- 开发测试时使用 `local` 模式，利用 Web 界面进行配置和调试
- 生产部署时使用 `normal` 模式，连接云平台的 gRPC 服务器

如需切换模式，修改配置文件后重启驱动即可。

### Q: 配置文件保存在哪里？

A: 配置自动保存到 `data.json` 文件（路径可通过 `dataConfig` 配置项自定义），支持热加载和持久化。

### Q: 如何配置消息队列？

A: 在 `config.yaml` 中配置 `mq` 节点，支持 `mqtt`、`rabbit`、`kafka`、`local` 四种类型。

### Q: 数据点如何处理范围校验？

A: 在 `Tag` 配置中设置 `Range` 参数，SDK 会自动进行范围校验和异常值处理。

### Q: 如何调试驱动？

A: 实现 `Debug` 方法，或使用 `app.LogDebug()` 输出调试日志。

## 相关链接

- [官网](https://airiot.cn)
- [文档中心](https://docs.airiot.cn)
- [Go开发文档](https://docs.airiot.cn/category/go)

## 许可证

请参考项目根目录的 LICENSE 文件。