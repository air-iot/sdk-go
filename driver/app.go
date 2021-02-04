package driver

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/air-iot/sdk-go/api"
	"github.com/air-iot/sdk-go/conn/mqtt"
	"github.com/air-iot/sdk-go/conn/rabbit"
	"github.com/air-iot/sdk-go/conn/websocket"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/air-iot/sdk-go/logger"
)

type App interface {
	Start(driver Driver, handlers ...Handler)
	WritePoints(point Point) error
	LogDebug(uid string, msg interface{})
	LogInfo(uid string, msg interface{})
	LogWarn(uid string, msg interface{})
	LogError(uid string, msg interface{})
	GetLogger() *logrus.Logger
	ApiClient() api.Client
	// ConvertValue(tag, raw interface{}) (map[string]interface{}, interface{}, error)
}

type Driver interface {
	Start(App, []byte) error
	Reload(App, []byte) error
	Run(App, string, []byte) (interface{}, error)
	Debug(App, []byte) (interface{}, error)
	Stop(App) error
	Schema(App) (string, error)
}

type Handler interface {
	Start()
	Stop()
}

// DG 数据采集类
type app struct {
	*logrus.Logger
	sendMethod  string
	mqtt        *mqtt.Mqtt
	rabbit      *rabbit.Amqp
	ws          *websocket.Conn
	api         api.Client
	driverId    string
	serviceId   string
	driverName  string
	distributed string
	host        string
	port        int
}

// 存储数据
type Point struct {
	Uid      string  `json:"uid"`     // 设备编号
	ModelId  string  `json:"modelId"` // 模型id
	NodeId   string  `json:"nodeId"`  // 资产id
	Fields   []Field `json:"fields"`  // 数据点
	UnixTime int64   `json:"time"`    // 数据采集时间 毫秒数
}

// 字段
type Field struct {
	Tag   interface{} `json:"tag"`   // 数据点
	Value interface{} `json:"value"` // 数据采集值
}

type pointTmp struct {
	Uid      string                 `json:"uid"`
	ModelId  string                 `json:"modelId"`
	NodeId   string                 `json:"nodeId"`
	Fields   map[string]interface{} `json:"fields"`
	UnixTime int64                  `json:"time"`
}

type wsRequest struct {
	RequestId string      `json:"requestId"`
	Action    string      `json:"action"`
	Data      interface{} `json:"data"`
}

type wsResponse struct {
	RequestId string `json:"requestId"`
	Data      result `json:"data"`
}

type result struct {
	Code   int         `json:"code"`
	Result interface{} `json:"result"`
}

type command struct {
	NodeId  string      `json:"nodeId"`
	Command interface{} `json:"command"`
}

type resultMsg struct {
	Message string `json:"message"`
}

func init() {
	// 设置随机数种子
	rand.Seed(time.Now().Unix())
	runtime.GOMAXPROCS(runtime.NumCPU())

	viper.SetDefault("log.level", "INFO")
	viper.SetDefault("host", "traefik")
	viper.SetDefault("port", 80)

	viper.SetDefault("mqtt.host", "mqtt")
	viper.SetDefault("mqtt.port", 1883)
	viper.SetDefault("mqtt.username", "admin")
	viper.SetDefault("mqtt.password", "public")

	viper.SetDefault("rabbit.host", "rabbit")
	viper.SetDefault("rabbit.port", 5672)
	viper.SetDefault("rabbit.username", "admin")
	viper.SetDefault("rabbit.password", "public")

	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./etc/")
	if err := viper.ReadInConfig(); err != nil {
		//log.Println("读取配置,", err.Error())
		//os.Exit(1)
		logrus.Fatalln("读取配置,", err.Error())
	}
}

// NewDG 创建DG
func NewApp() App {
	var (
		host        = viper.GetString("host")
		port        = viper.GetInt("port")
		ak          = viper.GetString("credentials.ak")
		sk          = viper.GetString("credentials.sk")
		driverId    = viper.GetString("driver.id")
		driverName  = viper.GetString("driver.name")
		distributed = viper.GetString("driver.distributed")
		sendMethod  = viper.GetString("driver.sendMethod")
		logLevel    = viper.GetString("log.level")
	)
	if driverId == "" || driverName == "" {
		panic("驱动id或name不能为空")
	}
	a := new(app)
	a.Logger = logger.NewLogger(logLevel)
	a.distributed = distributed
	a.driverId = driverId
	a.driverName = driverName
	a.sendMethod = sendMethod
	a.serviceId = GetRandomString(20)
	mqttCli, err := mqtt.NewMqtt(viper.GetString("mqtt.host"), viper.GetInt("mqtt.port"), viper.GetString("mqtt.username"), viper.GetString("mqtt.password"))
	if err != nil {
		logrus.Fatalln("连接mqtt错误,", err.Error())
	}
	a.mqtt = mqttCli
	if sendMethod == "rabbit" {
		rabbitCli, err := rabbit.NewAmqp(viper.GetString("rabbit.host"), viper.GetInt("rabbit.port"), ak, sk, "/")
		if err != nil {
			//panic(err)
			logrus.Fatalln("连接rabbit错误,", err.Error())
		}
		a.rabbit = rabbitCli
	}
	a.host = host
	a.port = port
	a.api = api.NewClient("http", host, port, ak, sk)

	return a
}

// Start 开始服务
func (p *app) Start(driver Driver, handlers ...Handler) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	for _, handler := range handlers {
		handler.Start()
	}
	var wsConnected = false
	var reloadFlag = false
	go func() {
		var timeConnect = 0
		var timeOut = 10
		for {
			var err error
			p.ws, err = websocket.DialWS(fmt.Sprintf(`ws://%s:%d/driver/ws?driverId=%s&driverName=%s&serviceId=%s&distributed=%s`, p.host, p.port, p.driverId, p.driverName, p.serviceId, p.distributed))
			if err != nil {
				timeConnect++
				if timeConnect > 5 {
					timeOut = 60
				}
				logrus.Errorf("尝试重新连接WebSocket第 %d 次失败,%s", timeConnect, err.Error())
				time.Sleep(time.Second * time.Duration(timeOut))
				continue
			}
			var handler = func() {
				for {
					var msg1 = new(wsRequest)
					err := p.ws.ReadJSON(&msg1)
					if err != nil {
						p.Logger.Warnln("服务端关闭,", err.Error())
						return
					}

					var r result
					switch msg1.Action {
					case "start":
						reloadFlag = true
						c, err := p.api.DriverConfig(p.driverId, p.serviceId)
						if err != nil {
							r = result{Code: http.StatusBadRequest, Result: resultMsg{Message: fmt.Sprintf("查询配置错误,%s", err.Error())}}
						} else {
							if err := driver.Start(p, c); err != nil {
								r = result{Code: http.StatusBadRequest, Result: resultMsg{Message: err.Error()}}
							} else {
								r = result{Code: http.StatusOK, Result: resultMsg{Message: "驱动启动成功"}}
							}
						}
					case "reload":
						reloadFlag = true
						c, err := p.api.DriverConfig(p.driverId, p.serviceId)
						if err != nil {
							r = result{Code: http.StatusBadRequest, Result: resultMsg{Message: fmt.Sprintf("查询配置错误,%s", err.Error())}}
						} else {
							if err := driver.Reload(p, c); err != nil {
								r = result{Code: http.StatusBadRequest, Result: resultMsg{Message: err.Error()}}
							} else {
								r = result{Code: http.StatusOK, Result: resultMsg{Message: "驱动重启成功"}}
							}
						}
					case "run":
						cmdByte, _ := json.Marshal(msg1.Data)
						cmd := new(command)
						err := json.Unmarshal(cmdByte, cmd)
						if err != nil {
							r = result{Code: http.StatusBadRequest, Result: resultMsg{Message: fmt.Sprintf("指令转换错误,%s", err.Error())}}
						} else {
							cmdByte, _ := json.Marshal(cmd.Command)
							if res1, err := driver.Run(p, cmd.NodeId, cmdByte); err != nil {
								r = result{Code: http.StatusBadRequest, Result: resultMsg{Message: err.Error()}}
							} else {
								if res1 == nil {
									res1 = resultMsg{"指令写入成功"}
								}
								r = result{Code: http.StatusOK, Result: res1}
							}
						}
					case "debug":
						debugByte, err := json.Marshal(msg1.Data)
						if err != nil {
							r = result{Code: http.StatusBadRequest, Result: resultMsg{Message: fmt.Sprintf("序列化数据错误,%s", err.Error())}}
						} else {
							if r1, err := driver.Debug(p, debugByte); err != nil {
								r = result{Code: http.StatusBadRequest, Result: resultMsg{Message: err.Error()}}
							} else {
								r = result{Code: http.StatusOK, Result: r1}
							}
						}
					case "schema":
						if r1, err := driver.Schema(p); err != nil {
							r = result{Code: http.StatusBadRequest, Result: resultMsg{Message: err.Error()}}
						} else {
							r = result{Code: http.StatusOK, Result: r1}
						}
					default:
						r = result{Code: http.StatusBadRequest, Result: resultMsg{Message: "未找到执行动作"}}
					}
					err = p.ws.WriteJSON(&wsResponse{RequestId: msg1.RequestId, Data: r})
					if err != nil {
						p.Logger.Warnln("写数据错误,", err.Error())
						return
					}
				}
			}
			timeConnect = 0
			timeOut = 10
			wsConnected = true
			handler()
		}
	}()
	//if p.distributed == "" {
	go func() {
		var c1 = make([]byte, 0)
		for {
			if wsConnected {
				c, err := p.api.DriverConfig(p.driverId, p.serviceId)
				if err != nil {
					p.Logger.Warnln("查询配置错误,", err.Error())
					//time.Sleep(time.Second * 10)
					//continue
				} else {
					c1 = c
				}
				break
			} else {
				time.Sleep(time.Second * 10)
			}
		}
		if !reloadFlag {
			if err := driver.Start(p, c1); err != nil {
				p.Logger.Warnln("驱动启动错误,", err.Error())
			}
		}
	}()
	//}
	sig := <-ch
	close(ch)
	if err := driver.Stop(p); err != nil {
		p.Logger.Warnln("驱动停止,", err.Error())
	}
	p.stop()
	for _, handler := range handlers {
		handler.Stop()
	}
	p.Logger.Debugln("关闭服务,", sig)
	os.Exit(0)
}

// GetLogger 获取日志
func (p *app) GetLogger() *logrus.Logger {
	return p.Logger
}

// Stop 服务停止
func (p *app) stop() {
	if p.ws != nil {
		p.ws.Close()
	}
	p.mqtt.Close()
	if p.sendMethod == "rabbit" {
		p.rabbit.Close()
	}
}

// SendMsg 发送消息
func (p *app) send(topic string, b []byte) error {
	if token := p.mqtt.Publish(topic, 0, false, string(b)); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// WritePoints 写数据点数据
func (p *app) WritePoints(point Point) error {
	if point.Uid == "" || point.NodeId == "" || point.ModelId == "" || point.Fields == nil || len(point.Fields) == 0 {
		return errors.New("数据有空值")
	}
	fields := make(map[string]interface{})
	for _, field := range point.Fields {
		if field.Tag == nil || field.Value == nil {
			p.Logger.Warnf("资产 [%s] 数据点为空", point.Uid)
			continue
		}
		tagByte, err := json.Marshal(field.Tag)
		if err != nil {
			p.Logger.Warnf("资产 [%s] 数据点序列化错误: %s", point.Uid, err.Error())
			continue
		}

		tag := new(Tag)
		err = json.Unmarshal(tagByte, tag)
		if err != nil {
			p.Logger.Warnf("资产 [%s] 数据点序列化tag结构体错误: %s", point.Uid, err.Error())
			continue
		}

		var value decimal.Decimal
		//vType := reflect.TypeOf(raw).String()
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
			fields[tag.ID] = field.Value
			continue
		}

		val := p.convertRange(tag.Range, p.convertValue(tag, value))
		if val != nil {
			fields[tag.ID] = val
		}
	}

	if len(fields) == 0 {
		return errors.New("数据点为空值")
	}
	b, err := json.Marshal(&pointTmp{Uid: point.Uid, ModelId: point.ModelId, NodeId: point.NodeId, UnixTime: point.UnixTime, Fields: fields})
	if err != nil {
		return err
	}
	p.Logger.Debugf("保存数据,%s", string(b))
	if p.sendMethod == "rabbit" {
		return p.rabbit.Send("data", fmt.Sprintf("data.%s", point.Uid), b)
	} else {
		return p.mqtt.Send(fmt.Sprintf("data/%s", point.Uid), string(b))
	}
}

// active fixed  boundary  discard
func (p *app) convertRange(tagRange *Range, raw decimal.Decimal) (val *float64) {
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
func (p *app) convertValue(tagTemp *Tag, raw decimal.Decimal) (val decimal.Decimal) {
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

// LogDebug 写日志数据
func (p *app) Log(topic string, msg interface{}) {
	l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "message": msg}
	b, err := json.Marshal(l)
	if err != nil {
		return
	}
	err = p.send("logs/"+topic, b)
	if err != nil {
		return
	}
}

// LogDebug 写日志数据
func (p *app) LogDebug(uid string, msg interface{}) {
	l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "uid": uid, "message": msg}
	b, err := json.Marshal(l)
	if err != nil {
		return
	}
	err = p.send("logs/debug/"+uid, b)
	if err != nil {
		return
	}
}

// LogInfo 写日志数据
func (p *app) LogInfo(uid string, msg interface{}) {
	l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "uid": uid, "message": msg}
	b, err := json.Marshal(l)
	if err != nil {
		return
	}
	err = p.send("logs/info/"+uid, b)
	if err != nil {
		return
	}
}

// LogWarn 写日志数据
func (p *app) LogWarn(uid string, msg interface{}) {
	l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "uid": uid, "message": msg}
	b, err := json.Marshal(l)
	if err != nil {
		return
	}
	err = p.send("logs/warn/"+uid, b)
	if err != nil {
		return
	}
}

// LogError 写日志数据
func (p *app) LogError(uid string, msg interface{}) {
	l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "uid": uid, "message": msg}
	b, err := json.Marshal(l)
	if err != nil {
		return
	}
	err = p.send("logs/error/"+uid, b)
	if err != nil {
		return
	}
}

// ApiClient api 接口客户端
func (p *app) ApiClient() api.Client {
	return p.api
}
