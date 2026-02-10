package driver

import (
	"context"
	"net/http"

	pb "github.com/air-iot/api-client-go/v4/driver"
	"github.com/gin-gonic/gin"

	"github.com/air-iot/sdk-go/v4/driver/entity"
)

// Driver 驱动接口，定义了驱动实现必须实现的方法
type Driver interface {
	// Schema 返回驱动配置的 Schema 定义
	// 用于动态生成驱动配置界面
	// 参数:
	//   - locale: 国际化语言代码，如 "zh"、"en"
	// 返回:
	//   - schema: 驱动配置的 JSON Schema 格式字符串
	//   - err: 错误信息
	Schema(ctx context.Context, app App, locale string) (schema string, err error)

	// Start 驱动启动方法
	// 用于初始化驱动连接、启动数据采集等
	// 参数:
	//   - driverConfig: 驱动配置数据（JSON 格式），包含实例、模型及设备信息
	// 返回:
	//   - err: 启动失败时返回错误信息
	Start(ctx context.Context, app App, driverConfig []byte) (err error)

	// RegisterRoutes 注册自定义 HTTP 路由
	// 在 HTTP 服务启动前调用，驱动可以注册自定义路由
	// 参数:
	//   - router: Gin 路由引擎实例
	RegisterRoutes(router *gin.RouterGroup)

	// Run 执行单设备指令
	// 向设备下发控制指令或配置命令
	// 参数:
	//   - command: 指令参数，包含 table（表标识）、id（设备编号）、serialNo（流水号）、command（指令内容）
	// 返回:
	//   - result: 指令执行结果，格式自定义
	//   - err: 执行失败时返回错误信息
	Run(ctx context.Context, app App, command *entity.Command) (result interface{}, err error)

	// BatchRun 批量执行多设备指令
	// 一次性向多个设备下发相同的指令
	// 参数:
	//   - command: 批量指令参数，包含 table（表标识）、ids（设备编号列表）、serialNo（流水号）、command（指令内容）
	// 返回:
	//   - result: 批量指令执行结果，格式自定义
	//   - err: 执行失败时返回错误信息
	BatchRun(ctx context.Context, app App, command *entity.BatchCommand) (result interface{}, err error)

	// WriteTag 写入数据点
	// 向设备写入指定的数据点值
	// 参数:
	//   - command: 写入参数，包含 table（表标识）、id（设备编号）、serialNo（流水号）、command（数据点内容）
	// 返回:
	//   - result: 写入操作结果，格式自定义
	//   - err: 写入失败时返回错误信息
	WriteTag(ctx context.Context, app App, command *entity.Command) (result interface{}, err error)

	// Debug 调试驱动
	// 用于驱动开发调试，可返回设备连接状态、测试数据等
	// 参数:
	//   - debugConfig: 调试参数（JSON 格式），内容自定义
	// 返回:
	//   - result: 调试结果，格式自定义
	//   - err: 调试失败时返回错误信息
	Debug(ctx context.Context, app App, debugConfig []byte) (result interface{}, err error)

	// HttpProxy HTTP 代理接口
	// 用于处理驱动自定义的 HTTP 请求
	// 参数:
	//   - t: 请求接口标识
	//   - header: HTTP 请求头
	//   - data: 请求体数据
	// 返回:
	//   - result: 响应结果，格式自定义
	//   - err: 处理失败时返回错误信息
	HttpProxy(ctx context.Context, app App, t string, header http.Header, data []byte) (result interface{}, err error)

	// ConfigUpdate 配置更新回调
	// 当驱动配置在平台侧被修改时触发
	// 参数:
	//   - data: 配置更新请求，包含变更内容
	// 返回:
	//   - err: 更新失败时返回错误信息
	ConfigUpdate(ctx context.Context, app App, data *pb.ConfigUpdateRequest) (err error)

	// Stop 驱动停止方法
	// 用于清理资源、关闭连接等
	// 返回:
	//   - err: 停止失败时返回错误信息
	Stop(ctx context.Context, app App) (err error)
}
