package driver

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"syscall"
	"time"

	"github.com/air-iot/sdk-go/api"
	"github.com/air-iot/sdk-go/conn/mqtt"
	"github.com/air-iot/sdk-go/conn/rabbit"
	"github.com/air-iot/sdk-go/conn/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/air-iot/sdk-go/logger"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type App interface {
	Start(driver Driver, handlers ...Handler)
	WritePoints(point Point) error
	LogDebug(uid string, msg interface{})
	LogInfo(uid string, msg interface{})
	LogWarn(uid string, msg interface{})
	LogError(uid string, msg interface{})
	GetLogger() *logrus.Logger
	//ConvertValue(tag, raw interface{}) (map[string]interface{}, interface{}, error)
}

type Driver interface {
	Start(App, []byte) error
	Reload(App, []byte) error
	Run(App, string, []byte) error
	Debug(App, []byte) (interface{}, error)
	Stop(App) error
	Schema() string
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
		log.Println("读取配置,", err.Error())
		os.Exit(1)
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
	a.Logger.Infoln(1, ak)
	a.distributed = distributed
	a.driverId = driverId
	a.driverName = driverName
	a.sendMethod = sendMethod
	a.serviceId = GetRandomString(10)
	mqttCli, err := mqtt.NewMqtt(viper.GetString("mqtt.host"), viper.GetInt("mqtt.port"), ak, sk)
	if err != nil {
		panic(err)
	}
	a.mqtt = mqttCli
	if sendMethod == "rabbit" {
		rabbitCli, err := rabbit.NewAmqp(viper.GetString("rabbit.host"), viper.GetInt("rabbit.port"), ak, sk, "/")
		if err != nil {
			panic(err)
		}
		a.rabbit = rabbitCli
	}

	wsCli, err := websocket.DialWS(fmt.Sprintf(`ws://%s:%d/driver/ws?driverId=%s&driverName=%s&serviceId=%s&distributed=%s`, host, port, driverId, driverName, a.serviceId, distributed))
	if err != nil {
		panic(err)
	}
	a.ws = wsCli
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
	go func() {
		for {
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
							if err := driver.Run(p, cmd.NodeId, cmdByte); err != nil {
								r = result{Code: http.StatusBadRequest, Result: resultMsg{Message: err.Error()}}
							} else {
								r = result{Code: http.StatusOK, Result: resultMsg{"指令写入成功"}}
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
			handler()
		}
	}()
	if p.distributed == "" {
		go func() {
			var c1 []byte
			for {
				c, err := p.api.DriverConfig(p.driverId, p.serviceId)
				if err != nil {
					p.Logger.Warnln("查询配置错误,", err.Error())
					time.Sleep(time.Second * 5)
					continue
				}
				c1 = c
				break
			}
			if err := driver.Start(p, c1); err != nil {
				p.Logger.Panic("驱动启动错误,", err.Error())
			}
		}()
	}
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
	p.ws.Close()
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
		if field.Tag == nil {
			continue
		}

		tag, val, err := p.ConvertValue(field.Tag, field.Value)
		if err != nil {
			p.Logger.Warnln("转换值错误,", err)
			continue
		}

		if id1, ok := tag["id"]; ok {
			if id, ok := id1.(string); ok {
				fields[id] = val
			}
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

// ConvertValue 数据点值转换
func (p *app) ConvertValue(tagTemp, raw interface{}) (tag map[string]interface{}, val interface{}, err error) {

	tag = make(map[string]interface{})

	b, err := json.Marshal(tagTemp)
	if err != nil {
		return tag, raw, err
	}
	err = json.Unmarshal(b, &tag)
	if err != nil {
		return tag, raw, err
	}

	var value float64
	vType := reflect.TypeOf(raw).String()

	switch vType {
	case "float32", "float64":
		value = reflect.ValueOf(raw).Float()
	case "uint", "uintptr", "uint8", "uint16", "uint32", "uint64":
		value = float64(reflect.ValueOf(raw).Uint())
	case "int", "int8", "int16", "int32", "int64":
		value = float64(reflect.ValueOf(raw).Int())
	default:
		return tag, raw, nil
	}

	const (
		minValueKey = "minValue"
		maxValueKey = "maxValue"
		minRawKey   = "minRaw"
		maxRawKey   = "maxRaw"
	)

	// let { minValue, maxValue, minRaw, maxRaw }
	tagValue := make(map[string]float64)
	if val, ok := tag["tagValue"]; ok {
		if tagValueMap, ok := val.(map[string]interface{}); ok {
			p.convertValue(&tagValue, minValueKey, tagValueMap)
			p.convertValue(&tagValue, maxValueKey, tagValueMap)
			p.convertValue(&tagValue, minRawKey, tagValueMap)
			p.convertValue(&tagValue, maxRawKey, tagValueMap)
		}
	}
	if minRaw, ok := tagValue[minRawKey]; ok && value < minRaw {
		value = minRaw
	}

	if maxRaw, ok := tagValue[maxRawKey]; ok && value > maxRaw {
		value = maxRaw
	}

	minValue, ok1 := tagValue[minValueKey]
	maxValue, ok2 := tagValue[maxValueKey]
	minRaw, ok3 := tagValue[minRawKey]
	maxRaw, ok4 := tagValue[maxRawKey]
	if ok1 && ok2 && ok3 && ok4 && (maxRaw != minRaw) {
		value = (((value - minRaw) / (maxRaw - minRaw)) * (maxValue - minValue)) + minValue
	}

	if fixed, ok := tag["fixed"]; ok {
		switch val1 := fixed.(type) {
		case float64:
			if n2, err := strconv.ParseFloat(fmt.Sprintf("%."+strconv.Itoa(int(val1))+"f", value), 64); err == nil {
				value = n2
			}
		case int:
			if n2, err := strconv.ParseFloat(fmt.Sprintf("%."+strconv.Itoa(val1)+"f", value), 64); err == nil {
				value = n2
			}
		case string:
			if n2, err := strconv.ParseFloat(fmt.Sprintf("%."+val1+"f", value), 64); err == nil {
				value = n2
			}
		}

	}

	if mod, ok := tag["mod"]; ok {
		switch val1 := mod.(type) {
		case float64:
			value = value * val1
		case int:
			value = value * float64(val1)
		case string:
			if v, err := strconv.ParseFloat(val1, 64); err == nil {
				value = value * v
			}
		}
	}

	return tag, value, nil
}

func (p *app) convertValue(tagValue *map[string]float64, key string, tagValueMap map[string]interface{}) {
	if val, ok := tagValueMap[key]; ok {
		switch val1 := val.(type) {
		case float64:
			(*tagValue)[key] = val1
		case int:
			(*tagValue)[key] = float64(val1)
		case string:
			if v, err := strconv.ParseFloat(val1, 64); err == nil {
				(*tagValue)[key] = v
			}
		}
	}
}
