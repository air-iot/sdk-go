package driver

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/air-iot/json"
	"github.com/air-iot/logger"
	"github.com/air-iot/sdk-go/v4/driver/license"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/air-iot/sdk-go/v4/conn/mq"
	"github.com/air-iot/sdk-go/v4/driver/convert"
	"github.com/air-iot/sdk-go/v4/driver/entity"
	"github.com/air-iot/sdk-go/v4/utils/decrypt"
	"github.com/air-iot/sdk-go/v4/utils/numberx"
)

type App interface {
	Start(Driver)
	GetProjectId() string
	GetGroupID() string
	GetServiceId() string
	GetMQ() mq.MQ
	GetRouter() *gin.Engine
	StartHTTPServer() error
	WritePoints(context.Context, entity.Point) error
	SavePoints(ctx context.Context, tableId string, data *entity.WritePoint) error
	WriteEvent(context.Context, entity.Event) error
	WriteWarning(context.Context, entity.Warn) error
	WriteWarningRecovery(ctx context.Context, tableId, dataId string, w entity.WarnRecovery) error
	FindDevice(ctx context.Context, table, id string, ret interface{}) error
	RunLog(context.Context, entity.Log) error
	UpdateTableData(ctx context.Context, table, id string, custom map[string]interface{}) error
	LogDebug(table, id string, msg interface{})
	LogInfo(table, id string, msg interface{})
	LogWarn(table, id string, msg interface{})
	LogError(table, id string, msg interface{})
	GetCommands(ctx context.Context, table, id string, ret interface{}) error
	UpdateCommand(ctx context.Context, id string, data entity.DriverInstruct) error
	saveDataConfig(config []byte) error
	BroadcastRealtimeData(tableId string, data *entity.WritePoint) error
}

const (
	String     = "string"
	Float      = "float"
	Integer    = "integer"
	Boolean    = "boolean"
	BooleanRaw = "boolean_raw"
)

// app 数据采集类
type app struct {
	mq                mq.MQ
	stopped           bool
	cli               *Client
	clean             func()
	driver            Driver
	httpServer        *http.Server
	httpClean         func()
	router            *gin.Engine
	dataConfigMutex   sync.Mutex
	dataConfigModTime time.Time        // data.json 最后修改时间（用于区分内部/外部修改）
	dataConfigSkip    bool             // 是否跳过文件变化监听（内部保存时设置为 true）
	dataConfigCache   *dataConfigCache // data.json 数据缓存

	cacheValue sync.Map

	// WebSocket 实时数据推送
	wsClients      map[*websocketConn]struct{}
	wsClientsMutex sync.RWMutex

	// 设备状态跟踪
	deviceStatus      sync.Map // key: "table:id", value: *deviceStatusRecord
	deviceStatusMutex sync.RWMutex
}

// dataConfigCache data.json 数据缓存
type dataConfigCache struct {
	data    map[string]interface{} // 解析后的 JSON 数据
	version int64                  // 版本号（用于检测变化）
}

// websocketConn WebSocket 连接封装
type websocketConn struct {
	conn   any        // 实际类型为 *websocket.Conn，使用 any 避免 import 循环
	mu     sync.Mutex // 保护并发写入
	typ    string     // 连接类型: data, tag, device, model, status
	table  string     // 订阅的表ID，为空表示订阅所有
	device string     // 订阅的设备ID，为空表示订阅所有
}

// deviceStatusRecord 设备状态记录
type deviceStatusRecord struct {
	mu            sync.Mutex // 互斥锁，保护并发访问
	lastSeen      int64      // 最后一次上数时间（毫秒时间戳）
	status        string     // 当前状态
	lastStatus    string     // 上一次状态（用于检测变化）
	statusChanged bool       // 状态是否发生变化
}

func Init() {
	// 设置随机数种子
	//rand.Seed(time.Now().Unix())
	runtime.GOMAXPROCS(runtime.NumCPU())
	pflag.String("project", "default", "项目id")
	pflag.String("serviceId", "", "服务id")
	pflag.String("groupId", "", "组id")
	cfgPath := pflag.String("config", "./etc/", "配置文件")
	viper.SetDefault("log.level", 4)
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")

	// mq
	viper.SetDefault("mq.type", "mqtt")
	viper.SetDefault("mq.timeout", "60s")
	viper.SetDefault("mq.mqtt.schema", "tcp")
	viper.SetDefault("mq.mqtt.host", "mqtt")
	viper.SetDefault("mq.mqtt.port", 1883)
	viper.SetDefault("mq.mqtt.username", "admin")
	viper.SetDefault("mq.mqtt.password", "public")
	viper.SetDefault("mq.mqtt.keepAlive", 60)
	viper.SetDefault("mq.mqtt.connectTimeout", 20)
	viper.SetDefault("mq.mqtt.protocolVersion", 4)
	viper.SetDefault("mq.mqtt.tlsConfig.insecureSkipVerify", false)
	viper.SetDefault("mq.rabbit.host", "rabbit")
	viper.SetDefault("mq.rabbit.port", 5672)
	viper.SetDefault("mq.rabbit.username", "admin")
	viper.SetDefault("mq.rabbit.password", "public")
	viper.SetDefault("mq.kafka.brokers", []string{"kafka:9092"})

	// driver
	viper.SetDefault("driverGrpc.enable", true)
	viper.SetDefault("driverGrpc.host", "driver")
	viper.SetDefault("driverGrpc.port", 9224)
	viper.SetDefault("driverGrpc.health.requestTime", "10s")
	viper.SetDefault("driverGrpc.health.retry", 3)
	viper.SetDefault("driverGrpc.stream.heartbeat", "30s")
	viper.SetDefault("driverGrpc.waitTime", "5s")
	viper.SetDefault("driverGrpc.timeout", "600s")
	viper.SetDefault("driverGrpc.limit", 100)

	// etcd
	viper.SetDefault("etcd.endpoints", []string{"etcd:2379"})
	viper.SetDefault("etcd.dialTimeout", 60)
	viper.SetDefault("etcd.username", "root")
	viper.SetDefault("etcd.password", "")

	// etcd config
	viper.SetDefault("etcdConfig", "/airiot/config/pro.json")

	viper.SetDefault("http.enable", false)
	viper.SetDefault("http.mode", gin.ReleaseMode)
	viper.SetDefault("http.host", "0.0.0.0")
	viper.SetDefault("http.port", 8080)

	viper.SetDefault("dataFile.enable", false)
	viper.SetDefault("dataFile.path", "data.json")
	// api client
	viper.SetDefault("api.liteMode", false)
	viper.SetDefault("api.gateway", "http://127.0.0.1:3030/rest")
	viper.SetDefault("api.gatewayGrpc", "127.0.0.1:9224")
	viper.SetDefault("api.etcdConfig", "/airiot/config/pro.json")
	viper.SetDefault("api.metadata", map[string]string{"env": "local"})
	viper.SetDefault("api.type", "project")
	viper.SetDefault("api.projectId", "default")
	viper.SetDefault("api.ak", "")
	viper.SetDefault("api.sk", "")

	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	pflag.Parse()
	viper.AddConfigPath(*cfgPath)
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		panic(fmt.Errorf("解析命令行参数失败: %w；请检查启动参数格式（如 --config、--project）", err))
	}
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("读取配置文件失败: %w；请确认配置目录可访问，当前路径: %s", err, *cfgPath))
	}
	decrypt.Decode()
	if err := viper.Unmarshal(Cfg); err != nil {
		panic(fmt.Errorf("解析配置内容失败: %w；请检查配置文件语法和字段类型", err))
	}
}

// NewApp 创建App
func NewApp() App {
	Init()
	a := new(app)
	if Cfg.Driver.ID == "" || Cfg.Driver.Name == "" {
		panic("驱动配置无效: driver.id 和 driver.name 不能为空；请检查配置文件")
	}
	if Cfg.Project == "" {
		Cfg.Project = "default"
	}
	if Cfg.ServiceID == "" {
		Cfg.ServiceID = Cfg.Driver.ID + "-" + primitive.NewObjectID().Hex()
	}

	Cfg.Log.Syslog.ProjectId = Cfg.Project
	Cfg.Log.Syslog.ServiceName = fmt.Sprintf("%s-%s-%s", Cfg.Project, Cfg.ServiceID, Cfg.Driver.ID)
	logger.InitLogger(Cfg.Log)
	logger.Infof("启动配置已加载: project=%s, serviceId=%s, driverId=%s", Cfg.Project, Cfg.ServiceID, Cfg.Driver.ID)
	mqConn, clean, err := mq.NewMQ(Cfg.MQ)
	if err != nil {
		panic(fmt.Errorf("初始化消息队列失败: %w；请检查 mq.type、地址、端口和鉴权配置", err))
	}
	a.mq = mqConn
	a.clean = func() {
		clean()
	}
	a.cacheValue = sync.Map{}
	a.wsClients = make(map[*websocketConn]struct{})

	// 根据 datafile.enable 决定是否启动本地文件处理功能
	if Cfg.Datafile.Enable {
		// 启动 data 配置文件监听
		a.watchDataConfig()
	}

	// 如果配置了 HTTP，则初始化 router（pprof 会自动集成到 HTTP 服务中）
	if Cfg.HTTP.Enable && Cfg.HTTP.Host != "" && Cfg.HTTP.Port != "" {
		a.initRouter()
		// 启动设备状态检查器
		a.startDeviceStatusChecker()
	} else if Cfg.Pprof.Enable {
		// 只有在 HTTP 服务未启动时，才单独启动 pprof server
		go func() {
			addr := net.JoinHostPort(Cfg.Pprof.Host, Cfg.Pprof.Port)
			logger.Infof("pprof 服务已启动: addr=%s（可通过 go tool pprof 连接）", addr)
			if err := http.ListenAndServe(addr, nil); err != nil {
				logger.Errorf("pprof 服务启动失败: addr=%s, err=%v；请检查端口占用和网络绑定配置", addr, err)
				return
			}
		}()
	}

	return a
}

// loadDataConfigFromFile 从文件加载 data.json 配置（仅当 datafile.enable=true 时有效）
func (a *app) loadDataConfigFromFile() error {
	if !Cfg.Datafile.Enable {
		return nil
	}
	if Cfg.Datafile.Path == "" {
		return nil
	}
	data, err := os.ReadFile(Cfg.Datafile.Path)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Infof("未找到 data 配置文件，跳过加载: path=%s（如需启用请先创建文件）", Cfg.Datafile.Path)
			return nil
		}
		return fmt.Errorf("读取 data 配置文件失败: %w；请检查文件权限与路径: %s", err, Cfg.Datafile.Path)
	}
	if len(data) == 0 {
		return nil
	}

	// 获取文件修改时间
	info, _ := os.Stat(Cfg.Datafile.Path)
	a.dataConfigMutex.Lock()
	a.dataConfigModTime = info.ModTime()
	a.dataConfigMutex.Unlock()

	logger.Infof("开始加载 data 配置文件: path=%s", Cfg.Datafile.Path)

	// 更新内存缓存（供 HTTP 服务使用）
	if err := a.updateDataConfigCache(data); err != nil {
		logger.Warnf("更新 data 内存缓存失败: err=%v；配置已读取但 HTTP 查询结果可能不是最新", err)
	}

	if err := a.startDriverVerify(context.Background(), data); err != nil {
		return fmt.Errorf("根据 data 配置启动驱动失败: %w；请检查 data 内容与驱动参数", err)
	}
	// 注意：driver 可能为 nil（如果在 Start() 调用前被文件监听器触发）
	// 如果 driver 为 nil，配置会被加载到缓存，等待 Start() 调用时启动
	return nil
}

func (a *app) startDriverVerify(ctx context.Context, data []byte) error {
	if a.driver != nil {
		var free = map[string]string{
			"test":               "",
			"modbus":             "",
			"modbus_rtu":         "",
			"db-driver":          "",
			"driver-http-client": "",
			"driver-mqtt-client": "",
			"opcda":              "",
			"modbus_rtutcp":      "",
		}
		if _, ok := free[Cfg.Driver.ID]; !ok {
			ok, info, err := license.VerifyLicenseFromLib(Cfg.License, Cfg.Driver.ID, string(data))
			if err != nil {
				return fmt.Errorf("授权校验失败: %w；请检查 license 配置与驱动 ID", err)
			}
			logger.Infof("授权校验结果: %+v", info)
			if !ok {
				return fmt.Errorf("授权校验未通过: 点位数量超出授权范围；请减少点位数量或更新授权")
			}
		}
		if err := a.driver.Start(ctx, a, data); err != nil {
			return fmt.Errorf("驱动启动失败: %w；请检查驱动初始化参数与外部依赖", err)
		}
	}
	return nil
}

// loadDataAndStartDriver 加载配置文件并启动驱动（仅在 Start() 中调用，此时 driver 已确保非 nil）
func (a *app) loadDataAndStartDriver() error {
	if Cfg.Datafile.Path == "" {
		return nil
	}
	data, err := os.ReadFile(Cfg.Datafile.Path)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Infof("未找到 data 配置文件，跳过加载: path=%s（如需启用请先创建文件）", Cfg.Datafile.Path)
			return nil
		}
		return fmt.Errorf("读取 data 配置文件失败: %w；请检查文件权限与路径: %s", err, Cfg.Datafile.Path)
	}
	if len(data) == 0 {
		return nil
	}
	// 获取文件修改时间
	info, _ := os.Stat(Cfg.Datafile.Path)
	a.dataConfigMutex.Lock()
	a.dataConfigModTime = info.ModTime()
	a.dataConfigMutex.Unlock()

	logger.Infof("开始加载 data 配置并启动驱动: path=%s", Cfg.Datafile.Path)

	// 更新内存缓存（供 HTTP 服务使用）
	if err := a.updateDataConfigCache(data); err != nil {
		logger.Warnf("更新 data 内存缓存失败: err=%v；配置已读取但 HTTP 查询结果可能不是最新", err)
	}

	// 直接启动驱动
	ctx := context.Background()
	//if err := a.driver.Start(ctx, a, data); err != nil {
	//	return fmt.Errorf("启动驱动失败: %w", err)
	//}
	if err := a.startDriverVerify(ctx, data); err != nil {
		return fmt.Errorf("根据 data 配置启动驱动失败: %w；请检查 data 内容与驱动参数", err)
	}
	logger.Infof("已使用 data 配置启动驱动: path=%s", Cfg.Datafile.Path)
	return nil
}

// updateDataConfigCache 更新内存中的配置缓存（供 HTTP 服务查询使用）
func (a *app) updateDataConfigCache(config []byte) error {
	var configData map[string]interface{}
	if err := json.Unmarshal(config, &configData); err != nil {
		return fmt.Errorf("解析 data 配置失败: %w；请检查 JSON 格式和字段结构", err)
	}

	a.dataConfigMutex.Lock()
	if a.dataConfigCache == nil {
		a.dataConfigCache = &dataConfigCache{}
	}
	a.dataConfigCache.data = configData
	a.dataConfigCache.version++
	a.dataConfigMutex.Unlock()

	return nil
}

// saveDataConfig 保存配置到 data.json（如果 datafile.enable）并更新内存缓存
func (a *app) saveDataConfig(config []byte) error {
	// 始终更新内存缓存（供 HTTP 服务使用）
	if err := a.updateDataConfigCache(config); err != nil {
		logger.Warnf("更新 data 内存缓存失败: err=%v；配置已接收但 HTTP 查询结果可能不是最新", err)
	}

	// 只有启用 datafile 时才写文件
	if !Cfg.Datafile.Enable {
		return nil
	}
	if Cfg.Datafile.Path == "" {
		return nil
	}

	// 标记为内部修改，防止触发文件监听
	a.dataConfigMutex.Lock()
	a.dataConfigSkip = true
	a.dataConfigMutex.Unlock()

	if err := os.WriteFile(Cfg.Datafile.Path, config, 0644); err != nil {
		return fmt.Errorf("保存 data 配置文件失败: %w；请检查写入权限与磁盘空间，path=%s", err, Cfg.Datafile.Path)
	}

	// 更新修改时间
	info, _ := os.Stat(Cfg.Datafile.Path)
	a.dataConfigMutex.Lock()
	a.dataConfigModTime = info.ModTime()
	a.dataConfigMutex.Unlock()

	logger.Infof("data 配置文件已保存: path=%s", Cfg.Datafile.Path)
	return nil
}

// watchDataConfig 监听 data.json 文件变化（仅在 datafile.enable 时有效）
func (a *app) watchDataConfig() {
	if !Cfg.Datafile.Enable {
		return
	}
	if Cfg.Datafile.Path == "" {
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Errorf("创建文件监听器失败: err=%v；请检查系统文件句柄限制和路径权限", err)
		return
	}

	go func() {
		defer watcher.Close()
		for {
			// 检查文件是否存在
			if _, err := os.Stat(Cfg.Datafile.Path); err == nil {
				// 文件存在，开始监听
				if err := watcher.Add(Cfg.Datafile.Path); err == nil {
					logger.Infof("开始监听 data 配置文件: path=%s", Cfg.Datafile.Path)
					// 先读取一次文件并启动驱动
					if err := a.loadDataConfigFromFile(); err != nil {
						logger.Errorf("初始化加载 data 配置失败: path=%s, err=%v；请检查配置内容是否完整", Cfg.Datafile.Path, err)
					}
					break
				}
			}
			// 文件不存在或添加监听失败，等待后重试
			time.Sleep(5 * time.Second)
		}

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					a.dataConfigMutex.Lock()
					skip := a.dataConfigSkip
					a.dataConfigSkip = false
					a.dataConfigMutex.Unlock()

					if skip {
						// 内部修改，跳过
						continue
					}

					// 检查文件修改时间，避免重复处理
					info, err := os.Stat(Cfg.Datafile.Path)
					if err != nil {
						continue
					}

					a.dataConfigMutex.Lock()
					modTime := a.dataConfigModTime
					a.dataConfigMutex.Unlock()

					if info.ModTime().Equal(modTime) || info.ModTime().Before(modTime) {
						continue
					}

					logger.Infof("检测到 data 配置文件变化: path=%s", Cfg.Datafile.Path)
					if err := a.loadDataConfigFromFile(); err != nil {
						logger.Errorf("重新加载 data 配置失败: path=%s, err=%v；请修正配置后重试", Cfg.Datafile.Path, err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Errorf("文件监听异常: err=%v；建议检查文件系统事件是否正常", err)
			}
		}
	}()
}

// Start 开始服务
func (a *app) Start(driver Driver) {
	a.stopped = false
	a.driver = driver

	// 1. 如果启用 datafile，加载文件配置并启动驱动
	if Cfg.Datafile.Enable {
		// 尝试加载 data 配置文件并启动驱动
		if err := a.loadDataAndStartDriver(); err != nil {
			logger.Errorf("使用 data 配置启动驱动失败: err=%v；驱动将继续运行但 data 配置未生效", err)
		}
	}

	// 2. 如果启用 HTTP 服务，注册路由并启动
	if Cfg.HTTP.Enable {
		// 初始化内存缓存（HTTP 服务需要）
		a.dataConfigMutex.Lock()
		if a.dataConfigCache == nil {
			a.dataConfigCache = &dataConfigCache{
				data:    make(map[string]interface{}),
				version: 0,
			}
		}
		a.dataConfigMutex.Unlock()

		// 注册 driver 自定义路由
		if a.router != nil {
			driver.RegisterRoutes(a.GetRouter().Group(Cfg.Driver.ID))
			// 启动 HTTP 服务
			if err := a.StartHTTPServer(); err != nil {
				logger.Errorf("HTTP 服务启动失败: host=%s, port=%s, err=%v；请检查端口占用和监听地址配置", Cfg.HTTP.Host, Cfg.HTTP.Port, err)
			}
		}
	}

	cli := Client{cacheConfig: NewCacheConfig()}
	a.cli = cli.Start(a, driver)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	sig := <-ch
	close(ch)
	if err := driver.Stop(context.Background(), a); err != nil {
		logger.Warnf("驱动停止异常: err=%v；请检查 Stop 实现是否幂等并可安全退出", err)
	}
	cli.Stop()
	a.stop()
	logger.Debugf("收到退出信号，服务关闭完成: signal=%v", sig)
	os.Exit(0)
}

// Stop 服务停止
func (a *app) stop() {
	a.stopped = true
	if a.httpClean != nil {
		a.httpClean()
	}
	if a.clean != nil {
		a.clean()
	}
}

func (a *app) GetProjectId() string {
	return Cfg.Project
}

func (a *app) GetGroupID() string {
	return Cfg.GroupID
}

func (a *app) GetServiceId() string {
	return Cfg.ServiceID
}

func (a *app) GetMQ() mq.MQ {
	return a.mq
}

// WritePoints 写数据点数据
func (a *app) WritePoints(ctx context.Context, p entity.Point) error {
	//ctx = logger.NewModuleContext(ctx, entity.MODULE_WRITEPOINT)
	tableId := p.Table
	if tableId == "" {
		tableIdI, err := a.cli.cacheConfig.get(p.ID)
		if err != nil {
			return fmt.Errorf("获取设备所属表失败: %w；设备ID=%s，请确认设备已绑定到表", err, p.ID)
		}
		tableId = tableIdI
	}
	if p.ID == "" {
		return fmt.Errorf("写入数据点失败: 设备ID为空；请检查采集配置或上报参数")
	}
	if p.Fields == nil || len(p.Fields) == 0 {
		return fmt.Errorf("写入数据点失败: 字段列表为空；请检查采集点配置和上报值")
	}
	ctx = logger.NewTableContext(ctx, tableId)
	if Cfg.GroupID != "" {
		ctx = logger.NewGroupContext(ctx, Cfg.GroupID)
	}
	return a.writePoints(ctx, tableId, p)
}

func (a *app) writePoints(ctx context.Context, tableId string, p entity.Point) error {
	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), Cfg.MQ.Timeout)
	defer cancelTimeout()
	fields := make(map[string]interface{})
	newLogger := logger.WithContext(ctx)
	for _, field := range p.Fields {
		if field.Value == nil {
			newLogger.Warnf("数据点值为空，已跳过: table=%s, device=%s, tag=%s；请检查设备上报或点位映射", tableId, p.ID, field.Tag.ID)
			continue
		}
		tag := field.Tag
		if strings.TrimSpace(tag.ID) == "" {
			newLogger.Errorf("数据点标识为空，已跳过: table=%s, device=%s；请检查点位 ID 配置", tableId, p.ID)
			continue
		}

		var value decimal.Decimal
		switch valueTmp := field.Value.(type) {
		case float32:
			if math.IsNaN(float64(valueTmp)) || math.IsInf(float64(valueTmp), 0) {
				newLogger.Errorf("数据点值非法，已跳过: table=%s, device=%s, tag=%s, value=%v（NaN/Inf）；请检查设备原始数据或转换逻辑", tableId, p.ID, tag.ID, valueTmp)
				continue
			}
			value = decimal.NewFromFloat32(valueTmp)
		case float64:
			if math.IsNaN(valueTmp) || math.IsInf(valueTmp, 0) {
				newLogger.Errorf("数据点值非法，已跳过: table=%s, device=%s, tag=%s, value=%v（NaN/Inf）；请检查设备原始数据或转换逻辑", tableId, p.ID, tag.ID, valueTmp)
				continue
			}
			value = decimal.NewFromFloat(valueTmp)
		case uint:
			value = decimal.NewFromInt(int64(valueTmp))
		case uint8:
			value = decimal.NewFromInt(int64(valueTmp))
		case uint16:
			value = decimal.NewFromInt(int64(valueTmp))
		case uint32:
			value = decimal.NewFromInt(int64(valueTmp))
		case uint64:
			value = decimal.NewFromInt(int64(valueTmp))
		case int:
			value = decimal.NewFromInt(int64(valueTmp))
		case int8:
			value = decimal.NewFromInt(int64(valueTmp))
		case int16:
			value = decimal.NewFromInt(int64(valueTmp))
		case int32:
			value = decimal.NewFromInt32(valueTmp)
		case int64:
			value = decimal.NewFromInt(valueTmp)
		case []byte:
			fields[tag.ID] = fmt.Sprintf("hex__%s", hex.EncodeToString(valueTmp))
			continue
		default:
			valTmp, err := numberx.GetValueByType("", field.Value)
			if err != nil {
				errCtx := logger.NewErrorContext(ctx, err)
				logger.WithContext(errCtx).Errorf("数据点类型转换失败: table=%s, device=%s, tag=%s, value=%v, err=%v；请检查点位类型定义与上报值格式", tableId, p.ID, tag.ID, field.Value, err)
				continue
			}
			fields[tag.ID] = valTmp
			continue
		}
		val := convert.Value(&tag, value)
		if tag.Range != nil && (tag.Range.Enable == nil || *(tag.Range.Enable)) {
			cacheKey := fmt.Sprintf("%s__%s__%s", tableId, p.ID, tag.ID)
			preValF, ok := a.cacheValue.Load(cacheKey)
			var preVal *decimal.Decimal
			if ok {
				preF, ok := preValF.(*float64)
				if ok && preF != nil {
					preValue := decimal.NewFromFloat(*preF)
					preVal = &preValue
				}
			}
			newVal, rawVal, invalidType, save := convert.Range(tag.Range, preVal, &val)
			if newVal != nil {
				valTmp, err := numberx.GetValueByType("", newVal)
				if err != nil {
					errCtx := logger.NewErrorContext(ctx, err)
					logger.WithContext(errCtx).Errorf("范围转换后类型处理失败: table=%s, device=%s, tag=%s, value=%v, err=%v；请检查范围配置与目标类型", tableId, p.ID, tag.ID, newVal, err)
				} else {
					valTmp = convert.ValueFormat(&tag, valTmp)
					fields[tag.ID] = valTmp
					if save {
						a.cacheValue.Store(cacheKey, newVal)
					}
				}
			}
			if rawVal != nil {
				valTmp, err := numberx.GetValueByType("", rawVal)
				if err != nil {
					errCtx := logger.NewErrorContext(ctx, err)
					logger.WithContext(errCtx).Errorf("原始数据点类型转换失败: table=%s, device=%s, tag=%s, value=%v, err=%v；请检查原始值类型和点位定义", tableId, p.ID, tag.ID, rawVal, err)
				} else {
					fields[fmt.Sprintf("%s__invalid", tag.ID)] = valTmp
				}
			}
			if invalidType != "" {
				fields[fmt.Sprintf("%s__invalid__type", tag.ID)] = invalidType
			}
		} else {
			vTmp, _ := val.Float64()
			fields[tag.ID] = convert.ValueFormat(&tag, vTmp)
		}
	}
	if len(fields) == 0 {
		return fmt.Errorf("写入数据点失败: 所有字段均为空或无效；请检查采集值、点位映射及类型转换配置")
	}
	if p.UnixTime == 0 {
		p.UnixTime = time.Now().Local().UnixMilli()
	} else if p.UnixTime > 9999999999999 || p.UnixTime < 1000000000000 {
		return fmt.Errorf("时间戳无效: %d（应为 13 位毫秒时间戳，范围: 1000000000000-9999999999999）；请检查设备上报时间单位", p.UnixTime)
	}
	data := &entity.WritePoint{ID: p.ID, CID: p.CID, Source: "device", UnixTime: p.UnixTime, Fields: fields, FieldTypes: p.FieldTypes}
	//b, err := json.Marshal()
	//if err != nil {
	//	return err
	//}
	//return a.mq.Publish(ctxTimeout, []string{"data", Cfg.Project, tableId, p.ID}, b)
	return a.SavePoints(ctxTimeout, tableId, data)
}

func (a *app) SavePoints(ctx context.Context, tableId string, data *entity.WritePoint) error {
	if tableId == "" {
		return fmt.Errorf("保存数据点失败: 表ID为空；请检查设备表配置")
	}
	if data.ID == "" {
		return fmt.Errorf("保存数据点失败: 设备ID为空；请检查采集配置")
	}
	if len(data.Fields) == 0 {
		return fmt.Errorf("保存数据点失败: 字段列表为空；请检查采集点配置和上报值")
	}
	if data.Source == "" {
		data.Source = "device"
	}
	if data.UnixTime == 0 {
		data.UnixTime = time.Now().UnixMilli()
	} else if data.UnixTime > 9999999999999 || data.UnixTime < 1000000000000 {
		return fmt.Errorf("时间戳无效: %d（应为 13 位毫秒时间戳，范围: 1000000000000-9999999999999）；请检查设备上报时间单位", data.UnixTime)
	}
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化数据点失败: %w；请检查字段是否包含不可序列化类型", err)
	}
	if logger.IsLevelEnabled(logger.DebugLevel) {
		logger.Debugf("数据点已保存: table=%s, device=%s, payload=%s", tableId, data.ID, string(b))
	}
	// 广播到 WebSocket 客户端
	go a.BroadcastRealtimeData(tableId, data)
	// 更新设备状态
	a.updateDeviceLastSeen(tableId, data.ID)
	if err := a.mq.Publish(ctx, []string{"data", Cfg.Project, tableId, data.ID}, b); err != nil {
		return fmt.Errorf("发布数据点到 MQ 失败: %w；请检查 MQ 连接状态与主题权限", err)
	}
	return nil
}

func (a *app) WriteWarning(ctx context.Context, w entity.Warn) error {
	//ctx = logger.NewModuleContext(ctx, entity.MODULE_WARN)
	tableId := w.TableId
	if tableId == "" {
		tableIdI, err := a.cli.cacheConfig.get(w.TableDataId)
		if err != nil {
			return fmt.Errorf("获取设备所属表失败: %w；设备ID=%s，请确认设备已绑定到表", err, w.TableDataId)
		}
		tableId = tableIdI
	}
	w.TableId = tableId
	if w.TableDataId == "" {
		return fmt.Errorf("写入报警失败: 设备ID为空；请检查报警配置")
	}
	if tableId == "" {
		return fmt.Errorf("写入报警失败: 表ID为空；请检查报警配置")
	}
	ctx = logger.NewTableContext(ctx, tableId)
	if Cfg.GroupID != "" {
		ctx = logger.NewGroupContext(ctx, Cfg.GroupID)
	}
	if w.Time == nil {
		n := time.Now().Local()
		w.Time = &n
	}

	wt := entity.WarnSend{
		ID:          w.ID,
		Table:       entity.Table{ID: tableId},
		TableData:   entity.TableData{ID: w.TableDataId},
		Level:       w.Level,
		Ruleid:      w.Ruleid,
		Fields:      w.Fields,
		WarningType: w.WarningType,
		Processed:   w.Processed,
		Time:        w.Time.Format(time.RFC3339),
		Alert:       w.Alert,
		Status:      w.Status,
		Handle:      w.Handle,
		Desc:        w.Desc,
		I18nProp:    w.I18nProp,
	}
	b, err := json.Marshal(wt)
	if err != nil {
		return fmt.Errorf("序列化报警数据失败: %w；请检查报警字段格式", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), Cfg.MQ.Timeout)
	defer cancel()
	if err := a.mq.Publish(ctx, []string{"warningStorage", Cfg.Project, tableId, w.TableDataId}, b); err != nil {
		return fmt.Errorf("发布报警数据到 MQ 失败: %w；请检查 MQ 连接状态与主题权限", err)
	}
	return nil
}

// WriteWarningRecovery 报警恢复
func (a *app) WriteWarningRecovery(ctx context.Context, tableId, dataId string, w entity.WarnRecovery) error {
	//ctx = logger.NewModuleContext(ctx, entity.MODULE_WARN)
	if tableId == "" {
		return fmt.Errorf("报警恢复失败: 表ID为空；请检查报警恢复配置")
	}
	if dataId == "" {
		return fmt.Errorf("报警恢复失败: 设备ID为空；请检查报警恢复配置")
	}
	if len(w.ID) == 0 {
		return fmt.Errorf("报警恢复失败: 报警ID列表为空；请指定需要恢复的报警ID")
	}
	ctx = logger.NewTableContext(ctx, tableId)
	if Cfg.GroupID != "" {
		ctx = logger.NewGroupContext(ctx, Cfg.GroupID)
	}
	if w.Data.Time == nil {
		n := time.Now().Local()
		w.Data.Time = &n
	}
	wt := entity.WarnRecoverySend{
		ID: w.ID,
		Data: entity.WarnRecoveryDataSend{
			Time:   w.Data.Time.Format(time.RFC3339),
			Fields: w.Data.Fields,
		},
	}
	b, err := json.Marshal(wt)
	if err != nil {
		return fmt.Errorf("序列化报警恢复数据失败: %w；请检查恢复字段格式", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), Cfg.MQ.Timeout)
	defer cancel()
	if err := a.mq.Publish(ctx, []string{"warningUpdate", Cfg.Project, tableId, dataId}, b); err != nil {
		return fmt.Errorf("发布报警恢复数据到 MQ 失败: %w；请检查 MQ 连接状态与主题权限", err)
	}
	return nil
}

func (a *app) WriteEvent(ctx context.Context, event entity.Event) error {
	return a.cli.WriteEvent(ctx, event)
}

func (a *app) FindDevice(ctx context.Context, table, id string, ret interface{}) error {
	return a.cli.FindDevice(ctx, table, id, ret)
}

func (a *app) GetCommands(ctx context.Context, table, id string, ret interface{}) error {
	return a.cli.GetCommands(ctx, table, id, ret)
}

func (a *app) UpdateCommand(ctx context.Context, id string, data entity.DriverInstruct) error {
	return a.cli.UpdateCommand(ctx, id, data)
}

func (a *app) RunLog(ctx context.Context, l entity.Log) error {
	return a.cli.RunLog(ctx, l)
}

func (a *app) UpdateTableData(ctx context.Context, table, id string, custom map[string]interface{}) error {
	return a.cli.UpdateTableData(ctx, entity.TableData{
		TableID: table,
		ID:      id,
		Data:    custom,
	}, &map[string]interface{}{})
}

// Log 写日志数据
func (a *app) Log(topic string, msg interface{}) {
	l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "message": msg}
	b, err := json.Marshal(l)
	if err != nil {
		logger.Warnf("发送运行日志失败: topic=%s, err=%v；请检查日志字段是否可序列化", topic, err)
		return
	}
	if err := a.mq.Publish(context.Background(), []string{"logs", topic}, b); err != nil {
		logger.Warnf("发送运行日志失败: topic=%s, err=%v；请检查 MQ 连接状态与主题权限", topic, err)
		return
	}
}

// LogDebug 写日志数据
func (a *app) LogDebug(table, id string, msg interface{}) {
	if logger.IsLevelEnabled(logger.DebugLevel) {
		l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "message": msg}
		b, err := json.Marshal(l)
		if err != nil {
			logger.Warnf("发送调试日志失败: table=%s, device=%s, err=%v；请检查日志字段是否可序列化", table, id, err)
			return
		}
		if err := a.mq.Publish(context.Background(), []string{"logs", Cfg.Project, "debug", table, id}, b); err != nil {
			logger.Warnf("发送调试日志失败: table=%s, device=%s, err=%v；请检查 MQ 连接状态与主题权限", table, id, err)
			return
		}
	}
}

// LogInfo 写日志数据
func (a *app) LogInfo(table, id string, msg interface{}) {
	if logger.IsLevelEnabled(logger.InfoLevel) {
		l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "message": msg}
		b, err := json.Marshal(l)
		if err != nil {
			logger.Warnf("发送信息日志失败: table=%s, device=%s, err=%v；请检查日志字段是否可序列化", table, id, err)
			return
		}
		if err := a.mq.Publish(context.Background(), []string{"logs", Cfg.Project, "info", table, id}, b); err != nil {
			logger.Warnf("发送信息日志失败: table=%s, device=%s, err=%v；请检查 MQ 连接状态与主题权限", table, id, err)
			return
		}
	}
}

// LogWarn 写日志数据
func (a *app) LogWarn(table, id string, msg interface{}) {
	if logger.IsLevelEnabled(logger.WarnLevel) {
		l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "message": msg}
		b, err := json.Marshal(l)
		if err != nil {
			logger.Warnf("发送告警日志失败: table=%s, device=%s, err=%v；请检查日志字段是否可序列化", table, id, err)
			return
		}
		if err := a.mq.Publish(context.Background(), []string{"logs", Cfg.Project, "warn", table, id}, b); err != nil {
			logger.Warnf("发送告警日志失败: table=%s, device=%s, err=%v；请检查 MQ 连接状态与主题权限", table, id, err)
			return
		}
	}
}

// LogError 写日志数据
func (a *app) LogError(table, id string, msg interface{}) {
	if logger.IsLevelEnabled(logger.ErrorLevel) {
		l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "message": msg}
		b, err := json.Marshal(l)
		if err != nil {
			logger.Warnf("发送错误日志失败: table=%s, device=%s, err=%v；请检查日志字段是否可序列化", table, id, err)
			return
		}
		if err := a.mq.Publish(context.Background(), []string{"logs", Cfg.Project, "error", table, id}, b); err != nil {
			logger.Warnf("发送错误日志失败: table=%s, device=%s, err=%v；请检查 MQ 连接状态与主题权限", table, id, err)
			return
		}
	}
}

// BroadcastRealtimeData 广播实时数据到所有 WebSocket 客户端
func (a *app) BroadcastRealtimeData(tableId string, data *entity.WritePoint) error {
	a.wsClientsMutex.RLock()
	defer a.wsClientsMutex.RUnlock()

	if len(a.wsClients) == 0 {
		return nil
	}

	// 构建广播消息，包含 tableId
	msg := map[string]interface{}{
		"table":  tableId,
		"id":     data.ID,
		"fields": data.Fields,
		"time":   data.UnixTime,
	}

	// 序列化数据一次
	b, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化实时数据失败: %w；请检查字段是否包含不可序列化类型", err)
	}

	// 广播到所有匹配的客户端
	for client := range a.wsClients {
		// 按类型过滤：只有 type=data 的客户端才接收实时数据
		if client.typ != "data" {
			continue
		}
		// 按表过滤：如果客户端指定了表ID，只推送该表的数据
		if client.table != "" && client.table != tableId {
			continue
		}
		// 按设备过滤：如果客户端指定了设备ID，只推送该设备的数据；否则推送所有设备
		if client.device != "" && client.device != data.ID {
			continue
		}

		// 异步推送到客户端
		go func(c *websocketConn) {
			if err := c.sendData(b); err != nil {
				logger.Errorf("WebSocket 推送实时数据失败: type=%s, table=%s, device=%s, err=%v；连接将被移除，建议客户端重连", c.typ, c.table, c.device, err)
				a.unregisterWebSocketClient(c)
			}
		}(client)
	}

	return nil
}

// registerWebSocketClient 注册 WebSocket 客户端
func (a *app) registerWebSocketClient(conn *websocketConn) {
	a.wsClientsMutex.Lock()
	defer a.wsClientsMutex.Unlock()
	a.wsClients[conn] = struct{}{}
	logger.Infof("WebSocket 客户端已连接: type=%s, table=%s, device=%s, active=%d", conn.typ, conn.table, conn.device, len(a.wsClients))
}

// unregisterWebSocketClient 注销 WebSocket 客户端
func (a *app) unregisterWebSocketClient(conn *websocketConn) {
	a.wsClientsMutex.Lock()
	defer a.wsClientsMutex.Unlock()
	if _, ok := a.wsClients[conn]; ok {
		delete(a.wsClients, conn)
		logger.Infof("WebSocket 客户端已断开: type=%s, table=%s, device=%s, active=%d", conn.typ, conn.table, conn.device, len(a.wsClients))
	}
}

// sendData 发送数据到 WebSocket 客户端（线程安全）
func (ws *websocketConn) sendData(data []byte) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	conn := ws.conn.(interface{ WriteMessage(int, []byte) error })
	return conn.WriteMessage(1, data) // 1 = TextMessage
}

// updateDeviceLastSeen 更新设备最后一次上数时间（仅在 HTTP 服务启用时有效）
func (a *app) updateDeviceLastSeen(tableId, deviceId string) {
	if !Cfg.HTTP.Enable {
		return
	}

	key := tableId + ":" + deviceId

	// 获取或创建设备状态记录，初始状态为 offline
	record, _ := a.deviceStatus.LoadOrStore(key, &deviceStatusRecord{
		status:     string(entity.DeviceStatusOffline),
		lastStatus: string(entity.DeviceStatusOffline),
	})

	now := time.Now().UnixMilli()
	r := record.(*deviceStatusRecord)

	// 加锁保护并发访问
	r.mu.Lock()
	defer r.mu.Unlock()

	// 更新最后上数时间和状态
	r.lastSeen = now
	r.status = string(entity.DeviceStatusOnline)
	r.lastStatus = r.status

	// 直接推送状态更新
	a.BroadcastDeviceStatus(tableId, deviceId, entity.DeviceStatusOnline, now)
}

// getDeviceTimeout 获取设备超时配置（秒）
// 优先级: 设备 > 模型 > 驱动
// 从内存缓存读取，不依赖文件
func (a *app) getDeviceTimeout(tableId, deviceId string) int {
	// 从内存缓存读取配置
	a.dataConfigMutex.Lock()
	config := a.dataConfigCache
	a.dataConfigMutex.Unlock()

	if config == nil || config.data == nil {
		return 0
	}

	// 1. 优先查找设备级别配置: tables[].devices[].settings.network.timeout
	tables, ok := config.data["tables"].([]interface{})
	if ok {
		for _, t := range tables {
			table, ok := t.(map[string]interface{})
			if !ok {
				continue
			}
			tid, _ := table["id"].(string)
			if tid != tableId {
				continue
			}

			// 找到对应的表，查找设备
			devices, ok := table["devices"].([]interface{})
			if !ok {
				break
			}
			for _, d := range devices {
				device, ok := d.(map[string]interface{})
				if !ok {
					continue
				}
				did, _ := device["id"].(string)
				if did != deviceId {
					continue
				}

				// 找到对应的设备，查找 settings.network.timeout
				if timeout := getNetworkTimeout(device); timeout > 0 {
					return timeout
				}
			}
			break
		}
	}

	// 2. 查找模型级别配置: tables[].settings.network.timeout
	if tables != nil {
		for _, t := range tables {
			table, ok := t.(map[string]interface{})
			if !ok {
				continue
			}
			tid, _ := table["id"].(string)
			if tid == tableId {
				// 找到对应的表，查找 settings.network.timeout
				if timeout := getNetworkTimeout(table); timeout > 0 {
					return timeout
				}
				break
			}
		}
	}

	// 3. 查找驱动级别配置: device.settings.network.timeout
	if device, ok := config.data["device"].(map[string]interface{}); ok {
		if timeout := getNetworkTimeout(device); timeout > 0 {
			return timeout
		}
	}

	// 都没有配置
	return 0
}

// getNetworkTimeout 从配置对象中获取 network.timeout
func getNetworkTimeout(obj map[string]interface{}) int {
	device, ok := obj["device"].(map[string]interface{})
	if !ok {
		return 0
	}
	settings, ok := device["settings"].(map[string]interface{})
	if !ok {
		return 0
	}

	network, ok := settings["network"].(map[string]interface{})
	if !ok {
		return 0
	}

	timeout, ok := network["timeout"].(float64)
	if !ok {
		return 0
	}

	return int(timeout)
}

// checkDeviceStatus 检查所有设备状态并推送状态变化（仅在 HTTP 服务启用时有效）
func (a *app) checkDeviceStatus() {
	if !Cfg.HTTP.Enable {
		return
	}

	now := time.Now().UnixMilli()

	a.deviceStatus.Range(func(key, value interface{}) bool {
		record := value.(*deviceStatusRecord)
		keyStr := key.(string)

		// 解析 tableId 和 deviceId
		parts := strings.SplitN(keyStr, ":", 2)
		if len(parts) != 2 {
			return true
		}
		tableId := parts[0]
		deviceId := parts[1]

		// 获取该设备的超时配置（设备 > 模型 > 驱动）
		timeout := a.getDeviceTimeout(tableId, deviceId)
		if timeout <= 0 {
			// 未配置超时时间，跳过该设备
			return true
		}

		timeoutMs := int64(timeout * 1000) // 转换为毫秒

		// 加锁保护并发访问
		record.mu.Lock()
		defer record.mu.Unlock()

		// 计算距离最后一次上数的时间
		elapsed := now - record.lastSeen
		newStatus := record.status

		if elapsed > timeoutMs {
			// 超时，判定为掉线
			newStatus = string(entity.DeviceStatusOffline)
		} else {
			// 未超时，判定为在线
			newStatus = string(entity.DeviceStatusOnline)
		}

		// 检查状态是否变化
		if newStatus != record.lastStatus {
			record.status = newStatus
			record.statusChanged = true
			record.lastStatus = newStatus

			// 推送状态更新
			a.BroadcastDeviceStatus(tableId, deviceId, entity.DeviceStatus(newStatus), record.lastSeen)
		}

		return true
	})
}

// BroadcastDeviceStatus 广播设备状态到所有订阅 status 类型的 WebSocket 客户端
func (a *app) BroadcastDeviceStatus(tableId, deviceId string, status entity.DeviceStatus, lastSeen int64) {
	a.wsClientsMutex.RLock()
	defer a.wsClientsMutex.RUnlock()

	if len(a.wsClients) == 0 {
		return
	}

	// 构建广播消息
	msg := map[string]interface{}{
		"table":    tableId,
		"id":       deviceId,
		"status":   status,
		"lastSeen": lastSeen,
	}

	// 序列化数据
	b, err := json.Marshal(msg)
	if err != nil {
		logger.Errorf("序列化设备状态消息失败: table=%s, device=%s, err=%v；请检查状态字段是否可序列化", tableId, deviceId, err)
		return
	}

	// 广播到所有订阅 status 类型的客户端
	for client := range a.wsClients {
		if client.typ != "status" {
			continue
		}

		// 按表过滤：如果客户端指定了表ID，只推送该表的状态
		if client.table != "" && client.table != tableId {
			continue
		}

		// 按设备过滤：如果客户端指定了设备ID，只推送该设备的状态
		if client.device != "" && client.device != deviceId {
			continue
		}

		// 异步推送到客户端
		go func(c *websocketConn) {
			if err := c.sendData(b); err != nil {
				logger.Errorf("WebSocket 推送设备状态失败: type=%s, table=%s, device=%s, err=%v；连接将被移除，建议客户端重连", c.typ, c.table, c.device, err)
				a.unregisterWebSocketClient(c)
			}
		}(client)
	}
}

// startDeviceStatusChecker 启动设备状态检查定时器
func (a *app) startDeviceStatusChecker() {
	go func() {
		ticker := time.NewTicker(1 * time.Second) // 每1秒检查一次
		defer ticker.Stop()

		for range ticker.C {
			a.checkDeviceStatus()
		}
	}()
}

// getTableDevices 获取指定表下的所有设备ID
func (a *app) getTableDevices(tableId string) []string {
	if a.cli == nil || a.cli.cacheConfig == nil {
		return []string{}
	}

	var devices []string
	a.cli.cacheConfig.lock.RLock()
	defer a.cli.cacheConfig.lock.RUnlock()

	// 遍历所有设备，找到属于指定表的设备
	for deviceId, tables := range a.cli.cacheConfig.data {
		if _, ok := tables[tableId]; ok {
			devices = append(devices, deviceId)
		}
	}

	return devices
}
