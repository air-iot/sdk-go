# AirIoT SDK Go

English | [简体中文](./README.md)

AirIoT SDK Go is a Go language SDK for developing AirIoT platform extension services, providing various types of extension capabilities including device drivers, algorithm services, data relays, and flow plugins.

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Module Types](#module-types)
- [Usage Examples](#usage-examples)
- [Configuration](#configuration)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [FAQ](#faq)
- [Troubleshooting](#troubleshooting)
- [Requirements](#requirements)
- [Changelog](#changelog)
- [License](#license)

## Features

- **Device Driver**: Support data collection and control for devices with multiple protocols
- **Algorithm Service**: Provide custom algorithm execution capabilities
- **Data Relay**: Implement data forwarding and processing
- **Flow Plugin**: Support custom logic for flow nodes
- **Flow Extension**: Provide flow node schema configuration capabilities

## Quick Start

### Installation

```bash
go get github.com/air-iot/sdk-go/v4
```

### Basic Usage

```go
import "github.com/air-iot/sdk-go/v4/driver"

func main() {
    // Create application instance
    app := driver.NewApp()
    // Start driver
    app.Start(yourDriver)
}
```

## Module Types

### 1. Device Driver

Used for developing device driver programs to implement device data collection and control functions.

#### Driver Interface

| Method | Description | Parameters |
|--------|-------------|------------|
| `Schema` | Return JSON Schema definition for driver configuration, used to dynamically generate driver configuration UI | `ctx context.Context`<br>`app App`<br>`locale string` - Internationalization language code (e.g., "zh", "en") |
| `Start` | Start the driver, initialize device connections, start data collection, etc. | `ctx context.Context`<br>`app App`<br>`driverConfig []byte` - Driver configuration data (JSON format), including instance, model, and device information |
| `RegisterRoutes` | Register custom HTTP routes, called before HTTP service starts | `router *gin.Engine` - Gin router engine instance |
| `Run` | Execute single device command, send control command or configuration command to device | `ctx context.Context`<br>`app App`<br>`command *entity.Command` - Command parameters |
| `BatchRun` | Batch execute multiple device commands, send same command to multiple devices at once | `ctx context.Context`<br>`app App`<br>`command *entity.BatchCommand` - Batch command parameters |
| `WriteTag` | Write data point, write specified data point value to device | `ctx context.Context`<br>`app App`<br>`command *entity.Command` - Write parameters |
| `Debug` | Debug driver, return device connection status, test data, etc. | `ctx context.Context`<br>`app App`<br>`debugConfig []byte` - Debug parameters (JSON format) |
| `HttpProxy` | HTTP proxy interface, handle driver custom HTTP requests | `ctx context.Context`<br>`app App`<br>`t string` - Request interface identifier<br>`header http.Header` - HTTP request headers<br>`data []byte` - Request body data |
| `ConfigUpdate` | Configuration update callback, triggered when driver configuration is modified on platform side | `ctx context.Context`<br>`app App`<br>`data *pb.ConfigUpdateRequest` - Configuration update request |
| `Stop` | Stop driver, clean up resources, close connections, etc. | `ctx context.Context`<br>`app App` |

#### Command Parameter Structure

```go
// Command single device command parameters
type Command struct {
    Table    string `json:"table"`    // Table identifier
    Id       string `json:"id"`       // Device ID
    SerialNo string `json:"serialNo"` // Serial number
    Command  []byte `json:"command"`  // Command content
}

// BatchCommand batch device command parameters
type BatchCommand struct {
    Table    string   `json:"table"`    // Table identifier
    Ids      []string `json:"ids"`      // Device ID list
    SerialNo string   `json:"serialNo"` // Serial number
    Command  []byte   `json:"command"`  // Command content
}

// Command status constants
const (
    COMMAND_STATUS_TIMEOUT CommandStatus = "timeout" // Timeout
    COMMAND_STATUS_READY   CommandStatus = "ready"   // Pending
    COMMAND_STATUS_REVOKE  CommandStatus = "revoke"  // Revoked
    COMMAND_STATUS_SUCCESS CommandStatus = "success" // Success
    COMMAND_STATUS_FAIL    CommandStatus = "fail"    // Failed
)
```

#### App Interface

The App interface provided by the SDK for driver-platform interaction:

| Method | Description | Returns |
|--------|-------------|---------|
| `GetProjectId()` | Get project ID | `string` |
| `GetGroupID()` | Get group ID | `string` |
| `GetServiceId()` | Get service ID | `string` |
| `GetMQ()` | Get message queue instance | `mq.MQ` |
| `GetRouter()` | Get HTTP router group | `*gin.RouterGroup` |
| `StartHTTPServer()` | Start HTTP server | `error` |
| `WritePoints()` | Write measurement point data | `error` |
| `SavePoints()` | Save measurement point data to specified table | `error` |
| `WriteEvent()` | Write event | `error` |
| `WriteWarning()` | Write warning | `error` |
| `WriteWarningRecovery()` | Write warning recovery | `error` |
| `FindDevice()` | Query device information | `error` |
| `RunLog()` | Write run log | `error` |
| `UpdateTableData()` | Update table data | `error` |
| `LogDebug()` | Record debug log | - |
| `LogInfo()` | Record info log | - |
| `LogWarn()` | Record warning log | - |
| `LogError()` | Record error log | - |
| `GetCommands()` | Get pending commands | `error` |
| `UpdateCommand()` | Update command status | `error` |
| `BroadcastRealtimeData()` | Broadcast real-time data | `error` |

#### Data Types

```go
// Point storage data point
type Point struct {
    Table      string            `json:"table"`      // Table ID
    ID         string            `json:"id"`         // Device ID
    CID        string            `json:"cid"`        // Sub-device ID
    Fields     []Field           `json:"fields"`     // Data point list
    UnixTime   int64             `json:"time"`       // Data collection time (millisecond timestamp)
    FieldTypes map[string]string `json:"fieldTypes"` // Data point type mapping
}

// Field single data point field
type Field struct {
    Tag   Tag         `json:"tag"`   // Data point configuration
    Value interface{} `json:"value"` // Data collection value
}

// Event event data
type Event struct {
    Table    string                 `json:"table"`    // Table ID
    ID       string                 `json:"id"`       // Device ID
    Type     string                 `json:"type"`     // Event type
    Level    string                 `json:"level"`    // Event level
    Content  string                 `json:"content"`  // Event content
    UnixTime int64                  `json:"time"`     // Event time (millisecond timestamp)
    Data     map[string]interface{} `json:"data"`     // Additional data
}

// Warn warning data
type Warn struct {
    Table      string                 `json:"table"`      // Table ID
    ID         string                 `json:"id"`         // Device ID
    Type       string                 `json:"type"`       // Warning type
    Level      string                 `json:"level"`      // Warning level
    Content    string                 `json:"content"`    // Warning content
    UnixTime   int64                  `json:"time"`       // Warning time (millisecond timestamp)
    Data       map[string]interface{} `json:"data"`       // Additional data
    RecoverTag string                 `json:"recoverTag"` // Recovery tag
}
```

### 2. Algorithm Service

Used for developing custom algorithm services to provide algorithm execution capabilities.

#### Service Interface

| Method | Description | Parameters |
|--------|-------------|------------|
| `Schema` | Query JSON Schema for algorithm configuration, supports multiple languages | `ctx context.Context`<br>`app App`<br>`locale string` - Internationalization language code |
| `Start` | Start algorithm service, initialize algorithm resources | `ctx context.Context`<br>`app App` |
| `Run` | Execute algorithm, process input data and return results | `ctx context.Context`<br>`app App`<br>`input []byte` - Execution parameters, format: `{"function":"algorithmName","input":{}}` |
| `Stop` | Stop algorithm service, release resources | `ctx context.Context`<br>`app App` |

#### Input/Output Format

```go
// Algorithm execution input format
{
    "function": "algorithmName",  // Algorithm name
    "input": {                    // Algorithm input parameters, should match Schema defined format
        "param1": "value1",
        "param2": "value2"
    }
}

// Algorithm execution output format
{
    "result": "value",            // Algorithm calculation result
    "status": "success"           // Execution status
}
```

### 3. Data Relay

Used for implementing data forwarding and processing functions, can serve as a data bridge service connecting different data sources and targets.

#### DataRelay Interface

| Method | Description | Parameters |
|--------|-------------|------------|
| `Start` | Start data relay service | `ctx context.Context`<br>`app App`<br>`config []byte` - Configuration data (JSON format) |
| `HttpProxy` | HTTP proxy interface, handle custom HTTP requests | `ctx context.Context`<br>`app App`<br>`t string` - Request interface identifier<br>`header http.Header` - HTTP request headers<br>`data []byte` - Request data |

#### Use Cases

- Data format conversion
- Data forwarding and distribution
- Third-party system integration
- Data caching and queue processing

### 4. Flow Plugin

Used for developing custom logic for flow nodes, can be used in AirIoT platform's flow orchestration.

#### Flow Interface

| Method | Description | Parameters |
|--------|-------------|------------|
| `Handler` | Execute flow plugin, process flow node logic | `ctx context.Context`<br>`app App`<br>`request *Request` - Execution parameters |
| `Debug` | Debug flow plugin, return execution logs and results | `ctx context.Context`<br>`app App`<br>`request *DebugRequest` - Debug parameters |

#### Request Parameter Structure

```go
// Request flow execution request
type Request struct {
    ProjectId  string `json:"projectId,omitempty"`  // Project ID
    FlowId     string `json:"flowId,omitempty"`     // Flow ID
    Job        string `json:"job,omitempty"`        // Flow instance ID
    ElementId  string `json:"elementId,omitempty"`  // Node ID
    ElementJob string `json:"elementJob,omitempty"` // Node instance ID
    Config     []byte `json:"config,omitempty"`     // Node configuration
}

// DebugRequest flow debug request
type DebugRequest struct {
    ProjectId string `json:"projectId,omitempty"` // Project ID
    FlowId    string `json:"flowId,omitempty"`    // Flow ID
    ElementId string `json:"elementId,omitempty"` // Node ID
    Config    []byte `json:"config,omitempty"`    // Node configuration
}

// DebugResult debug result
type DebugResult struct {
    Logs  []Syslog               `json:"logs"`  // Debug log list
    Value map[string]interface{} `json:"value"` // Debug output value
}

// Syslog system log
type Syslog struct {
    Level string `json:"level"` // Log level
    Time  string `json:"time"`  // Log time
    Msg   string `json:"msg"`   // Log message
}
```

### 5. Flow Extension

Provide schema configuration and execution capabilities for flow nodes, used to create configurable flow extension services.

#### Extension Interface

| Method | Description | Parameters |
|--------|-------------|------------|
| `Schema` | Query JSON Schema definition for extension service, supports multiple languages | `ctx context.Context`<br>`app App`<br>`locale string` - Internationalization language code |
| `Run` | Execute extension service, process input data and return results | `ctx context.Context`<br>`app App`<br>`input []byte` - Execution parameters, should match Schema format |

#### Difference from Flow

- **Flow**: Used directly in flow orchestration, implements node logic through code
- **FlowExtension**: Provides configurable extension service, defines configuration UI through Schema, executes with configuration data as input

## Usage Examples

### Driver Complete Example

```go
package main

import (
    "context"
    "github.com/air-iot/sdk-go/v4/driver"
    "github.com/air-iot/sdk-go/v4/driver/entity"
)

type MyDriver struct{}

// Schema return driver configuration schema definition
func (d *MyDriver) Schema(ctx context.Context, app driver.App, locale string) (string, error) {
    return `{
        "type": "object",
        "properties": {
            "host": {
                "type": "string",
                "title": "Host Address",
                "default": "localhost"
            },
            "port": {
                "type": "integer",
                "title": "Port",
                "default": 502
            }
        }
    }`, nil
}

// Start start driver
func (d *MyDriver) Start(ctx context.Context, app driver.App, driverConfig []byte) error {
    // Parse configuration
    // Initialize device connections
    // Start data collection
    return nil
}

// Run execute single device command
func (d *MyDriver) Run(ctx context.Context, app driver.Driver, command *entity.Command) (interface{}, error) {
    // Parse command content
    // Send command to device
    // Return execution result
    return map[string]interface{}{"status": "success"}, nil
}

// BatchRun batch execute multiple device commands
func (d *MyDriver) BatchRun(ctx context.Context, app driver.Driver, command *entity.BatchCommand) (interface{}, error) {
    // Process commands for multiple devices
    return map[string]interface{}{"status": "success"}, nil
}

// WriteTag write data point
func (d *MyDriver) WriteTag(ctx context.Context, app driver.Driver, command *entity.Command) (interface{}, error) {
    // Write data point to device
    return map[string]interface{}{"status": "success"}, nil
}

// Debug debug driver
func (d *MyDriver) Debug(ctx context.Context, app driver.Driver, debugConfig []byte) (interface{}, error) {
    // Return debug information
    return map[string]interface{}{
        "connections": []string{"device1", "device2"},
        "status": "online",
    }, nil
}

// HttpProxy HTTP proxy interface
func (d *MyDriver) HttpProxy(ctx context.Context, app driver.Driver, t string, header http.Header, data []byte) (interface{}, error) {
    // Handle custom HTTP request
    return map[string]interface{}{"result": "ok"}, nil
}

// ConfigUpdate configuration update callback
func (d *MyDriver) ConfigUpdate(ctx context.Context, app driver.Driver, data *pb.ConfigUpdateRequest) error {
    // Handle configuration update
    return nil
}

// Stop stop driver
func (d *MyDriver) Stop(ctx context.Context, app driver.Driver) error {
    // Clean up resources, close connections
    return nil
}

// RegisterRoutes register custom HTTP routes
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

### Algorithm Complete Example

```go
package main

import (
    "context"
    "github.com/air-iot/sdk-go/v4/algorithm"
)

type MyAlgorithm struct{}

// Schema return algorithm configuration schema definition
func (a *MyAlgorithm) Schema(ctx context.Context, app algorithm.App, locale string) (string, error) {
    return `{
        "type": "object",
        "properties": {
            "threshold": {
                "type": "number",
                "title": "Threshold",
                "default": 100
            }
        }
    }`, nil
}

// Start start algorithm service
func (a *MyAlgorithm) Start(ctx context.Context, app algorithm.App) error {
    // Initialize algorithm resources
    return nil
}

// Run execute algorithm
func (a *MyAlgorithm) Run(ctx context.Context, app algorithm.App, input []byte) (interface{}, error) {
    // Parse input data
    // Execute algorithm logic
    // Return calculation result
    return map[string]interface{}{
        "result": 42,
        "status": "success",
    }, nil
}

// Stop stop algorithm service
func (a *MyAlgorithm) Stop(ctx context.Context, app algorithm.App) error {
    // Release resources
    return nil
}

func main() {
    app := algorithm.NewApp()
    app.Start(&MyAlgorithm{})
}
```

## Configuration

### Driver Configuration Example

```yaml
serviceId: your-service-id
project: your-project
driver:
  id: go-driver-demo
  name: Demo Driver
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

### Algorithm Configuration Example

```yaml
serviceId: your-service-id
algorithm:
  id: go-algorithm-demo
  name: Demo Algorithm
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

### Configuration Parameters

#### Driver Configuration Parameters

| Parameter | Description | Type | Default |
|-----------|-------------|------|---------|
| `serviceId` | Service ID, uniquely identifies service instance | string | Auto-generated |
| `project` | Project ID | string | - |
| `driver.id` | Driver identifier | string | - |
| `driver.name` | Driver name | string | - |
| `driverGrpc.enable` | Enable gRPC connection | bool | false |
| `driverGrpc.host` | gRPC server address | string | - |
| `driverGrpc.port` | gRPC server port | int | 9224 |
| `driverGrpc.healthRequestTime` | Health check request interval | duration | 10s |
| `driverGrpc.waitTime` | Reconnection wait time | duration | 5s |
| `log.level` | Log level (1-5) | int | 4 |
| `log.format` | Log format (json/text) | string | json |
| `mq.type` | Message queue type (local/mqtt/kafka) | string | local |
| `http.enable` | Enable HTTP service | bool | false |
| `http.host` | HTTP service listen address | string | 0.0.0.0 |
| `http.port` | HTTP service port | int | 8080 |
| `dataFile.enable` | Enable local data file | bool | false |
| `dataFile.path` | data.json file path | string | ./data.json |
| `pprof.enable` | Enable pprof profiling | bool | false |
| `pprof.host` | pprof listen address | string | - |
| `pprof.port` | pprof listen port | int | - |

#### Algorithm Configuration Parameters

| Parameter | Description | Type | Default |
|-----------|-------------|------|---------|
| `serviceId` | Service ID | string | Auto-generated |
| `algorithm.id` | Algorithm identifier | string | - |
| `algorithm.name` | Algorithm name | string | - |
| `algorithmGrpc.host` | gRPC server address | string | - |
| `algorithmGrpc.port` | gRPC server port | int | 9236 |
| `algorithmGrpc.limit` | Concurrency limit | int | 100 |
| `algorithm.timeout` | Execution timeout (seconds) | int | 600 |
| `api.gateway` | API gateway address | string | - |
| `api.gatewayGrpc` | API gRPC address | string | - |
| `api.type` | API type | string | project |
| `api.projectId` | Project ID | string | default |
| `api.ak` | Access Key | string | - |
| `api.sk` | Secret Key | string | - |

#### Log Level Description

| Level | Value | Description |
|-------|-------|-------------|
| Trace | 1 | Most detailed log information |
| Debug | 2 | Debug information |
| Info | 3 | General information |
| Warn | 4 | Warning information |
| Error | 5 | Error information |

#### Message Queue Configuration

##### MQTT Configuration

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

##### Kafka Configuration

```yaml
mq:
  type: kafka
  kafka:
    brokers:
      - localhost:9092
    topic: airiot-driver
```

## Examples

For complete example code, please refer to the [example](./example) directory:

- [Driver example](./example/driver) - Complete device driver example
- [Algorithm example](./example/algorithm) - Complete algorithm service example
- [Data relay example](./example/data_relay) - Complete data relay example
- [Flow example](./example/flow) - Complete flow plugin example
- [Flow extension example](./example/flow_extension) - Complete flow extension example

## Best Practices

### Driver Development Recommendations

1. **Connection Management**: Use connection pools to manage device connections, avoid frequent creation and destruction of connections
2. **Concurrency Control**: Implement concurrency control for device access to avoid simultaneous operations on the same device
3. **Error Handling**: Implement comprehensive error handling mechanisms and log detailed error information
4. **Resource Cleanup**: Properly release all resources in the `Stop` method
5. **Configuration Validation**: Validate configuration parameters in the `Start` method

### Data Reporting Recommendations

1. **Batch Reporting**: For high-frequency collected data, batch reporting is recommended to reduce network overhead
2. **Timestamp**: Use the actual device collection time, not the reporting time
3. **Data Types**: Correctly set data types to ensure data precision
4. **Exception Handling**: Mark and handle abnormal collected data

### Performance Optimization

1. **Async Processing**: Use goroutines for time-consuming operations
2. **Cache Strategy**: Reasonably use caching to reduce repeated calculations
3. **Connection Reuse**: Reuse network connections to reduce handshake overhead
4. **Resource Limits**: Set reasonable concurrency limits to avoid resource exhaustion

## FAQ

### Q: How to debug a driver?

A: You can debug a driver through the following ways:
1. Use the `Debug` method to return debug information
2. Enable HTTP service and add custom debug interfaces
3. Use `LogDebug`/`LogInfo` and other methods to record logs
4. Enable pprof for performance analysis

### Q: How does the driver handle configuration updates?

A: Implement the `ConfigUpdate` method, which will be automatically called when the configuration is modified on the platform side.

### Q: How to implement device reconnection?

A: Implement heartbeat mechanism and reconnection logic in the driver:
1. Periodically check device connection status
2. Automatically reconnect when disconnection is detected
3. Use `WriteWarning` to report device offline warnings

### Q: What to do when data reporting fails?

A: Check the following points:
1. Confirm MQ configuration is correct
2. Check network connection
3. View error logs to locate the problem
4. Consider implementing local caching, re-report after network recovery

### Q: How to handle command timeout?

A:
1. Set reasonable command execution timeout
2. Use context to control goroutine lifecycle
3. Update command status to `timeout` after timeout

## Troubleshooting

### Startup Failure

1. Check if configuration file format is correct
2. Confirm Go version meets requirements (>= 1.23)
3. View error information in logs
4. Verify dependency packages are completely installed

### Connection Failure

1. Confirm network connection is normal
2. Check firewall settings
3. Verify target address and port configuration
4. Check if authentication information is correct

### Performance Issues

1. Enable pprof to analyze performance bottlenecks
2. Check for goroutine leaks
3. Optimize data batch reporting strategy
4. Adjust concurrency limit parameters

## Technical Support

For questions or suggestions, please contact us through:

- Submit an Issue
- View project documentation
- Contact technical support team

## Requirements

- Go 1.23 or higher

## Changelog

### v4
- Refactored driver interface, added more features
- Optimized performance and stability
- Improved error handling mechanism
- Added more configuration options

## License

Please refer to the project license file for details.
