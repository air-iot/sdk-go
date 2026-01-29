package driver

import (
	"context"
	"encoding/hex"
	"errors"
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
	GetRouter() *gin.RouterGroup
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
	dataConfigModTime time.Time // data.json 最后修改时间（用于区分内部/外部修改）
	dataConfigSkip    bool      // 是否跳过文件变化监听（内部保存时设置为 true）

	cacheValue sync.Map
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
	viper.SetDefault("driverGrpc.host", "")
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
	viper.SetDefault("dataConfig", "./data.json")
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
		panic(fmt.Errorf("解析命令行参数失败: %w，请检查参数格式是否正确", err))
	}
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("读取配置文件失败: %w，请检查配置文件路径 %s 是否正确", err, *cfgPath))
	}
	decrypt.Decode()
	if err := viper.Unmarshal(Cfg); err != nil {
		panic(fmt.Errorf("解析配置内容失败: %w，请检查配置文件格式是否正确", err))
	}
}

// NewApp 创建App
func NewApp() App {
	Init()
	a := new(app)
	if Cfg.Driver.ID == "" || Cfg.Driver.Name == "" {
		panic("驱动配置错误: driver.id 和 driver.name 不能为空，请检查配置文件")
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
	logger.Infof("启动配置=%+v", *Cfg)
	mqConn, clean, err := mq.NewMQ(Cfg.MQ)
	if err != nil {
		panic(fmt.Errorf("初始化消息队列失败: %w，请检查 mq 配置是否正确", err))
	}
	a.mq = mqConn
	a.clean = func() {
		clean()
	}
	a.cacheValue = sync.Map{}
	// 启动 data 配置文件监听
	a.watchDataConfig()
	// 如果配置了 HTTP，则初始化 router（pprof 会自动集成到 HTTP 服务中）
	if Cfg.HTTP.Host != "" && Cfg.HTTP.Port != "" {
		a.initRouter()
	} else if Cfg.Pprof.Enable {
		// 只有在 HTTP 服务未启动时，才单独启动 pprof server
		go func() {
			addr := net.JoinHostPort(Cfg.Pprof.Host, Cfg.Pprof.Port)
			logger.Infof("pprof服务启动: 地址=%s", addr)
			if err := http.ListenAndServe(addr, nil); err != nil {
				logger.Errorf("pprof服务启动失败: 地址=%s, 错误=%v", addr, err)
				return
			}
		}()
	}
	return a
}

// loadDataConfig 加载 data.json 配置并调用 driver.Start
func (a *app) loadDataConfig() error {
	if Cfg.DataConfig == "" {
		return nil
	}

	data, err := os.ReadFile(Cfg.DataConfig)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Infof("data配置文件不存在，跳过加载: %s", Cfg.DataConfig)
			return nil
		}
		return fmt.Errorf("读取data配置文件失败: %w", err)
	}

	// 获取文件修改时间
	info, _ := os.Stat(Cfg.DataConfig)
	a.dataConfigMutex.Lock()
	a.dataConfigModTime = info.ModTime()
	a.dataConfigMutex.Unlock()

	logger.Infof("加载data配置文件: %s", Cfg.DataConfig)
	if a.driver != nil {
		ctx := context.Background()
		// data.json 中的数据已经是驱动格式，直接使用
		if err := a.driver.Start(ctx, a, data); err != nil {
			return fmt.Errorf("使用data配置启动驱动失败: %w", err)
		}
		logger.Infof("使用data配置启动驱动成功")
	}
	return nil
}

// saveDataConfig 保存配置到 data.json
func (a *app) saveDataConfig(config []byte) error {
	if Cfg.DataConfig == "" {
		return nil
	}

	// 标记为内部修改，防止触发文件监听
	a.dataConfigMutex.Lock()
	a.dataConfigSkip = true
	a.dataConfigMutex.Unlock()

	if err := os.WriteFile(Cfg.DataConfig, config, 0644); err != nil {
		return fmt.Errorf("保存data配置文件失败: %w", err)
	}

	// 更新修改时间
	info, _ := os.Stat(Cfg.DataConfig)
	a.dataConfigMutex.Lock()
	a.dataConfigModTime = info.ModTime()
	a.dataConfigMutex.Unlock()

	logger.Infof("保存data配置文件: %s", Cfg.DataConfig)
	return nil
}

// watchDataConfig 监听 data.json 文件变化
func (a *app) watchDataConfig() {
	if Cfg.DataConfig == "" {
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Errorf("创建文件监听器失败: %v", err)
		return
	}

	go func() {
		defer watcher.Close()
		for {
			// 检查文件是否存在
			if _, err := os.Stat(Cfg.DataConfig); err == nil {
				// 文件存在，开始监听
				if err := watcher.Add(Cfg.DataConfig); err == nil {
					logger.Infof("开始监听data配置文件: %s", Cfg.DataConfig)
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
					info, err := os.Stat(Cfg.DataConfig)
					if err != nil {
						continue
					}

					a.dataConfigMutex.Lock()
					modTime := a.dataConfigModTime
					a.dataConfigMutex.Unlock()

					if info.ModTime().Equal(modTime) || info.ModTime().Before(modTime) {
						continue
					}

					logger.Infof("检测到data配置文件变化: %s", Cfg.DataConfig)
					if err := a.loadDataConfig(); err != nil {
						logger.Errorf("加载变化的data配置失败: %v", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Errorf("文件监听错误: %v", err)
			}
		}
	}()
}

// Start 开始服务
func (a *app) Start(driver Driver) {
	a.stopped = false
	a.driver = driver

	// 尝试加载 data 配置文件
	if err := a.loadDataConfig(); err != nil {
		logger.Errorf("加载data配置文件失败: %v", err)
	}

	// 注册 driver 自定义路由
	if a.router != nil {
		driver.RegisterRoutes(a.router)
		// 启动 HTTP 服务
		if err := a.StartHTTPServer(); err != nil {
			logger.Errorf("HTTP服务启动失败: %v", err)
		}
	}

	cli := Client{cacheConfig: NewCacheConfig()}
	a.cli = cli.Start(a, driver)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	sig := <-ch
	close(ch)
	if err := driver.Stop(context.Background(), a); err != nil {
		logger.Warnf("驱动停止: %v", err.Error())
	}
	cli.Stop()
	a.stop()
	logger.Debugf("关闭服务: 信号=%v", sig)
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
			return fmt.Errorf("获取设备表ID失败: %w，设备ID=%s", err, p.ID)
		}
		tableId = tableIdI
	}
	if p.ID == "" {
		return fmt.Errorf("设备ID不能为空，请检查数据点采集配置")
	}
	if p.Fields == nil || len(p.Fields) == 0 {
		return fmt.Errorf("数据点字段列表为空，请检查是否正确配置了采集数据点")
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
			newLogger.Warnf("存数据点: 设备表=%s,设备=%s. 设备数据点值为空", tableId, p.ID)
			continue
		}
		tag := field.Tag
		if strings.TrimSpace(tag.ID) == "" {
			newLogger.Errorf("写入数据点失败: 设备表=%s, 设备=%s, 数据点标识为空，请检查数据点配置", tableId, p.ID)
			continue
		}

		var value decimal.Decimal
		switch valueTmp := field.Value.(type) {
		case float32:
			if math.IsNaN(float64(valueTmp)) || math.IsInf(float64(valueTmp), 0) {
				newLogger.Errorf("写入数据点失败: 设备表=%s, 设备=%s, 数据点=%s, 值=%f (值不是合法的数字，NaN或Inf)", tableId, p.ID, tag.ID, valueTmp)
				continue
			}
			value = decimal.NewFromFloat32(valueTmp)
		case float64:
			if math.IsNaN(valueTmp) || math.IsInf(valueTmp, 0) {
				newLogger.Errorf("写入数据点失败: 设备表=%s, 设备=%s, 数据点=%s, 值=%f (值不是合法的数字，NaN或Inf)", tableId, p.ID, tag.ID, valueTmp)
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
				logger.WithContext(errCtx).Errorf("数据点类型转换失败: 设备表=%s, 设备=%s, 数据点=%s, 原值=%v, 错误=%v", tableId, p.ID, tag.ID, field.Value, err)
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
					logger.WithContext(errCtx).Errorf("数据点类型转换失败: 设备表=%s, 设备=%s, 数据点=%s, 转换值=%v, 错误=%v", tableId, p.ID, tag.ID, newVal, err)
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
					logger.WithContext(errCtx).Errorf("原始数据点类型转换失败: 设备表=%s, 设备=%s, 数据点=%s, 原始值=%v, 错误=%v", tableId, p.ID, tag.ID, rawVal, err)
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
		return errors.New("所有数据点字段均为空或无效，无法写入数据")
	}
	if p.UnixTime == 0 {
		p.UnixTime = time.Now().Local().UnixMilli()
	} else if p.UnixTime > 9999999999999 || p.UnixTime < 1000000000000 {
		return fmt.Errorf("时间戳无效: %d (应为13位毫秒级时间戳，范围: 1000000000000-9999999999999)", p.UnixTime)
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
		return fmt.Errorf("表ID不能为空，请检查设备表配置")
	}
	if data.ID == "" {
		return fmt.Errorf("设备ID不能为空，请检查数据点采集配置")
	}
	if len(data.Fields) == 0 {
		return fmt.Errorf("数据点字段列表为空，请检查是否正确配置了采集数据点")
	}
	if data.Source == "" {
		data.Source = "device"
	}
	if data.UnixTime == 0 {
		data.UnixTime = time.Now().UnixMilli()
	} else if data.UnixTime > 9999999999999 || data.UnixTime < 1000000000000 {
		return fmt.Errorf("时间戳无效: %d (应为13位毫秒级时间戳，范围: 1000000000000-9999999999999)", data.UnixTime)
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if logger.IsLevelEnabled(logger.DebugLevel) {
		logger.Debugf("存数据点: 设备表=%s,设备=%s,数据=%s. 保存数据成功", tableId, data.ID, string(b))
	}
	return a.mq.Publish(ctx, []string{"data", Cfg.Project, tableId, data.ID}, b)
}

func (a *app) WriteWarning(ctx context.Context, w entity.Warn) error {
	//ctx = logger.NewModuleContext(ctx, entity.MODULE_WARN)
	tableId := w.TableId
	if tableId == "" {
		tableIdI, err := a.cli.cacheConfig.get(w.TableDataId)
		if err != nil {
			return fmt.Errorf("获取设备表ID失败: %w，设备ID=%s", err, w.TableDataId)
		}
		tableId = tableIdI
	}
	w.TableId = tableId
	if w.TableDataId == "" {
		return fmt.Errorf("设备ID不能为空，请检查报警配置")
	}
	if tableId == "" {
		return fmt.Errorf("表ID不能为空，请检查报警配置")
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
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), Cfg.MQ.Timeout)
	defer cancel()
	return a.mq.Publish(ctx, []string{"warningStorage", Cfg.Project, tableId, w.TableDataId}, b)
}

// WriteWarningRecovery 报警恢复
func (a *app) WriteWarningRecovery(ctx context.Context, tableId, dataId string, w entity.WarnRecovery) error {
	//ctx = logger.NewModuleContext(ctx, entity.MODULE_WARN)
	if tableId == "" {
		return fmt.Errorf("表ID不能为空，请检查报警恢复配置")
	}
	if dataId == "" {
		return fmt.Errorf("设备ID不能为空，请检查报警恢复配置")
	}
	if len(w.ID) == 0 {
		return fmt.Errorf("报警ID列表不能为空，请指定需要恢复的报警ID")
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
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), Cfg.MQ.Timeout)
	defer cancel()
	return a.mq.Publish(ctx, []string{"warningUpdate", Cfg.Project, tableId, dataId}, b)
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
		return
	}
	if err := a.mq.Publish(context.Background(), []string{"logs", topic}, b); err != nil {
		return
	}
}

// LogDebug 写日志数据
func (a *app) LogDebug(table, id string, msg interface{}) {
	if logger.IsLevelEnabled(logger.DebugLevel) {
		l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "message": msg}
		b, err := json.Marshal(l)
		if err != nil {
			return
		}
		if err := a.mq.Publish(context.Background(), []string{"logs", Cfg.Project, "debug", table, id}, b); err != nil {
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
			return
		}
		if err := a.mq.Publish(context.Background(), []string{"logs", Cfg.Project, "info", table, id}, b); err != nil {
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
			return
		}
		if err := a.mq.Publish(context.Background(), []string{"logs", Cfg.Project, "warn", table, id}, b); err != nil {
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
			return
		}
		if err := a.mq.Publish(context.Background(), []string{"logs", Cfg.Project, "error", table, id}, b); err != nil {
			return
		}
	}
}
