package driver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/air-iot/json"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/air-iot/logger"
	"github.com/air-iot/sdk-go/v4/conn/mq"
	"github.com/air-iot/sdk-go/v4/utils/numberx"
)

type App interface {
	Start(Driver)
	WritePoints(Point) error
	WriteEvent(context.Context, Event) error
	RunLog(context.Context, Log) error
	UpdateTableData(ctx context.Context, table, id string, custom map[string]interface{}) error
	LogDebug(table, id string, msg interface{})
	LogInfo(table, id string, msg interface{})
	LogWarn(table, id string, msg interface{})
	LogError(table, id string, msg interface{})
	GetProjectId() string
}

const (
	String  = "string"
	Float   = "float"
	Integer = "integer"
	Boolean = "boolean"
)

// app 数据采集类
type app struct {
	*logrus.Logger
	mq      mq.MQ
	stopped bool
	cli     *Client
	clean   func()
}

func init() {
	// 设置随机数种子
	rand.Seed(time.Now().Unix())
	runtime.GOMAXPROCS(runtime.NumCPU())
	pflag.String("project", "default", "项目id")
	pflag.String("serviceId", "", "服务id")
	cfgPath := pflag.String("config", "./etc/", "配置文件")
	pflag.Parse()
	viper.SetDefault("log.level", 4)
	viper.SetDefault("log.format", "text")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("mq.type", "mqtt")
	viper.SetDefault("mq.mqtt.host", "mqtt")
	viper.SetDefault("mq.mqtt.port", 1883)
	viper.SetDefault("mq.mqtt.username", "admin")
	viper.SetDefault("mq.mqtt.password", "public")
	viper.SetDefault("mq.mqtt.keepAlive", 60)
	viper.SetDefault("mq.mqtt.connectTimeout", 20)
	viper.SetDefault("mq.mqtt.protocolVersion", 4)
	viper.SetDefault("mq.rabbit.host", "rabbit")
	viper.SetDefault("mq.rabbit.port", 5672)
	viper.SetDefault("mq.rabbit.username", "admin")
	viper.SetDefault("mq.rabbit.password", "public")
	viper.SetDefault("driverGrpc.host", "driver")
	viper.SetDefault("driverGrpc.port", 9224)
	viper.SetDefault("driverGrpc.health.requestTime", 10)
	viper.SetDefault("driverGrpc.waitTime", 5)
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(*cfgPath)
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		log.Fatalln("读取命令行参数错误,", err.Error())
	}
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln("读取配置错误,", err.Error())
	}

	if err := viper.Unmarshal(C); err != nil {
		log.Fatalln("配置解析错误: ", err.Error())
	}
}

// NewApp 创建App
func NewApp() App {
	a := new(app)
	if _, err := logger.NewLogger(C.Log); err != nil {
		panic(fmt.Errorf("初始化日志错误,%s", err))
	}
	if C.ServiceID == "" {
		panic("服务id不能为空")
	}
	if C.Driver.ID == "" || C.Driver.Name == "" {
		panic("驱动id或name不能为空")
	}
	if C.DriverGrpc.Health.RequestTime == 0 {
		C.DriverGrpc.Health.RequestTime = 10
	}
	if C.DriverGrpc.Health.Retry == 0 {
		C.DriverGrpc.Health.Retry = 3
	}
	if C.DriverGrpc.WaitTime == 0 {
		C.DriverGrpc.WaitTime = 5
	}
	logger.Debugf("配置: %+v", *C)
	mqConn, clean, err := mq.NewMQ(C.MQ)
	if err != nil {
		panic(fmt.Errorf("初始化消息队列错误,%s", err))
	}
	a.mq = mqConn
	a.clean = func() {
		clean()
	}
	return a
}

// Start 开始服务
func (a *app) Start(driver Driver) {
	a.stopped = false
	cli := Client{cacheConfig: sync.Map{}, cacheConfigNum: sync.Map{}}
	a.cli = cli.Start(a, driver)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	sig := <-ch
	close(ch)
	if err := driver.Stop(a); err != nil {
		logger.Warnln("驱动停止,", err.Error())
	}
	cli.Stop()
	a.stop()
	logger.Debugln("关闭服务,", sig)
	os.Exit(0)
}

// Stop 服务停止
func (a *app) stop() {
	a.stopped = true
	if a.clean != nil {
		a.clean()
	}
}

func (a *app) GetProjectId() string {
	return C.Project
}

// WritePoints 写数据点数据
func (a *app) WritePoints(p Point) error {
	tableId := p.Table
	if tableId == "" {
		tableIdI, ok := a.cli.cacheConfig.Load(p.ID)
		if !ok {
			return fmt.Errorf("传入表id为空且未在配置中找到")
		}
		devI, ok := a.cli.cacheConfigNum.Load(p.ID)
		if ok {
			devM, _ := devI.(map[string]interface{})
			if len(devM) >= 2 {
				return fmt.Errorf("传入表id为空且在配置中找到多个表id")
			}
		}
		tableId = tableIdI.(string)
	}
	if p.ID == "" {
		return fmt.Errorf("记录id为空")
	}
	if p.Fields == nil || len(p.Fields) == 0 {
		return fmt.Errorf("采集数据有空值")
	}
	return a.writePoints(context.Background(), tableId, p)
}

func (a *app) writePoints(ctx context.Context, tableId string, p Point) error {
	fields := make(map[string]interface{})
	for _, field := range p.Fields {
		if field.Tag == nil || field.Value == nil {
			logger.Warnf("表 %s 资产 %s 数据点为空", tableId, p.ID)
			continue
		}
		tagByte, err := json.Marshal(field.Tag)
		if err != nil {
			logger.Warnf("表 %s 资产 %s 数据点序列化错误: %s", tableId, p.ID, err.Error())
			continue
		}

		tag := new(Tag)
		err = json.Unmarshal(tagByte, tag)
		if err != nil {
			logger.Warnf("表 %s 资产 %s 数据点序列化tag结构体错误: %s", tableId, p.ID, err.Error())
			continue
		}
		var value decimal.Decimal
		switch valueTmp := field.Value.(type) {
		case float32:
			value = decimal.NewFromFloat32(valueTmp)
		case float64:
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
		default:
			valTmp, err := numberx.GetValueByType("", field.Value)
			if err != nil {
				logger.Errorf("表 %s 资产 %s 数据点转类型错误: %v", tableId, p.ID, err)
				continue
			}
			fields[tag.ID] = valTmp
			continue
		}

		val := a.convertRange(tag.Range, a.convertValue(tag, value))
		if val != nil {
			valTmp, err := numberx.GetValueByType("", val)
			if err != nil {
				logger.Errorf("表 %s 资产 %s 数据点转类型错误: %v", tableId, p.ID, err)
				continue
			}
			fields[tag.ID] = valTmp
		}
	}

	if len(fields) == 0 {
		return errors.New("数据点为空值")
	}
	if p.UnixTime == 0 {
		p.UnixTime = time.Now().Local().UnixNano() / 1e6
	}
	b, err := json.Marshal(&point{ID: p.ID, CID: p.CID, Source: "device", UnixTime: p.UnixTime, Fields: fields, FieldTypes: p.FieldTypes})
	if err != nil {
		return err
	}
	logger.Debugf("保存数据,%s", string(b))
	return a.mq.Publish(ctx, []string{"data", C.Project, tableId, p.ID}, b)
	//return nil
}

func (a *app) WriteEvent(ctx context.Context, event Event) error {
	return a.cli.WriteEvent(ctx, event)
}

func (a *app) RunLog(ctx context.Context, l Log) error {
	return a.cli.RunLog(ctx, l)
}

func (a *app) UpdateTableData(ctx context.Context, table, id string, custom map[string]interface{}) error {
	return a.cli.UpdateTableData(ctx, TableData{
		TableID: table,
		ID:      id,
		Data:    custom,
	}, &map[string]interface{}{})
}

// active fixed  boundary  discard
func (a *app) convertRange(tagRange *Range, raw decimal.Decimal) (val *float64) {
	value, _ := raw.Float64()
	if tagRange == nil {
		return &value
	}
	if tagRange.MinValue == nil || tagRange.MaxValue == nil || tagRange.Active == nil {
		return &value
	}
	minValue := decimal.NewFromFloat(*tagRange.MinValue)
	maxValue := decimal.NewFromFloat(*tagRange.MaxValue)
	if raw.GreaterThanOrEqual(minValue) && raw.LessThanOrEqual(maxValue) {
		return &value
	}
	switch *tagRange.Active {
	case "fixed":
		if tagRange.FixedValue == nil {
			return &value
		}
		return tagRange.FixedValue
	case "boundary":
		if raw.LessThan(minValue) {
			return tagRange.MinValue
		}
		if raw.GreaterThan(maxValue) {
			return tagRange.MaxValue
		}
	case "discard":
		return nil
	default:
		return &value
	}
	return &value
}

// ConvertValue 数据点值转换
func (a *app) convertValue(tagTemp *Tag, raw decimal.Decimal) (val decimal.Decimal) {
	var value = raw
	if tagTemp.TagValue != nil {
		if tagTemp.TagValue.MinRaw != nil {
			minRaw := decimal.NewFromFloat(*tagTemp.TagValue.MinRaw)
			if value.LessThan(minRaw) {
				value = minRaw
			}
		}

		if tagTemp.TagValue.MaxRaw != nil {
			maxRaw := decimal.NewFromFloat(*tagTemp.TagValue.MaxRaw)
			if value.GreaterThan(maxRaw) {
				value = maxRaw
			}
		}

		if tagTemp.TagValue.MinRaw != nil && tagTemp.TagValue.MaxRaw != nil && tagTemp.TagValue.MinValue != nil && tagTemp.TagValue.MaxValue != nil {
			//value = (((rawTmp - minRaw) / (maxRaw - minRaw)) * (maxValue - minValue)) + minValue
			minRaw := decimal.NewFromFloat(*tagTemp.TagValue.MinRaw)
			maxRaw := decimal.NewFromFloat(*tagTemp.TagValue.MaxRaw)
			minValue := decimal.NewFromFloat(*tagTemp.TagValue.MinValue)
			maxValue := decimal.NewFromFloat(*tagTemp.TagValue.MaxValue)
			if !maxRaw.Equal(minRaw) {
				value = raw.Sub(minRaw).Div(maxRaw.Sub(minRaw)).Mul(maxValue.Sub(minValue)).Add(minValue)
			}
		}
	}

	if tagTemp.Fixed != nil {
		value = value.Round(*tagTemp.Fixed)
	}

	if tagTemp.Mod != nil {
		value = value.Mul(decimal.NewFromFloat(*tagTemp.Mod))
	}

	return value
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
	l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "message": msg}
	b, err := json.Marshal(l)
	if err != nil {
		return
	}
	if err := a.mq.Publish(context.Background(), []string{"logs", C.Project, "debug", table, id}, b); err != nil {
		return
	}
}

// LogInfo 写日志数据
func (a *app) LogInfo(table, id string, msg interface{}) {
	l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "message": msg}
	b, err := json.Marshal(l)
	if err != nil {
		return
	}
	if err := a.mq.Publish(context.Background(), []string{"logs", C.Project, "info", table, id}, b); err != nil {
		return
	}
}

// LogWarn 写日志数据
func (a *app) LogWarn(table, id string, msg interface{}) {
	l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "message": msg}
	b, err := json.Marshal(l)
	if err != nil {
		return
	}
	if err := a.mq.Publish(context.Background(), []string{"logs", C.Project, "warn", table, id}, b); err != nil {
		return
	}
	return
}

// LogError 写日志数据
func (a *app) LogError(table, id string, msg interface{}) {
	l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "message": msg}
	b, err := json.Marshal(l)
	if err != nil {
		return
	}
	if err := a.mq.Publish(context.Background(), []string{"logs", C.Project, "error", table, id}, b); err != nil {
		return
	}
	return
}
