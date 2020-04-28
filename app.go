package sdk

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
	consulApi "github.com/hashicorp/consul/api"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"gopkg.in/resty.v1"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type App interface {
	Start(driver Driver, handlers ...Handler)
	WritePoints(point Point) error
	LogDebug(uid string, msg interface{})
	LogInfo(uid string, msg interface{})
	LogWarn(uid string, msg interface{})
	LogError(uid string, msg interface{})
	ConvertValue(tagTemp, raw interface{}) (interface{}, error)
}

type Driver interface {
	Start(App, []byte) error
	Reload(App, []byte) error
	Run(App, string, []byte) error
	Debug(App, []byte) (interface{}, error)
	Stop(App) error
}

type Handler interface {
	Start()
	Stop()
}

// DG 数据采集类
type app struct {
	dataAction     string
	mqtt           MQTT.Client
	rabbit         *amqp.Connection
	consul         *consulApi.Client
	ws             *websocket.Conn
	driverId       string
	driverName     string
	serviceId      string
	serviceName    string
	traefikAddress string
}

type Point struct {
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
	viper.SetDefault("traefik.host", "traefik")
	viper.SetDefault("traefik.port", 80)

	viper.SetDefault("consul.host", "consul")
	viper.SetDefault("consul.port", 8500)

	viper.SetDefault("data.action", "mqtt")

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
	viper.SetConfigType("ini")
	viper.SetConfigName("config")
	viper.AddConfigPath("./etc/")
	if err := viper.ReadInConfig(); err != nil {
		log.Println("读取配置,", err.Error())
		os.Exit(1)
	}
	logInit()
}

// NewDG 创建DG
func NewApp() App {
	var (
		driverId   = viper.GetString("driver.id")
		driverName = viper.GetString("driver.name")
	)
	if driverId == "" || driverName == "" {
		logrus.Panic("驱动id或name不能为空")
	}
	a := new(app)
	a.driverId = driverId
	a.driverName = driverName
	a.traefikAddress = fmt.Sprintf("%s:%d", viper.GetString("traefik.host"), viper.GetInt("traefik.port"))
	a.dataAction = viper.GetString("data.action")
	a.consulInit()
	a.mqttInit()
	a.rabbitInit()
	return a
}

func (p *app) consulInit() {
	cc := consulApi.DefaultConfig()
	cc.Address = fmt.Sprintf("%s:%d", viper.GetString("consul.host"), viper.GetInt("consul.port"))
	client, err := consulApi.NewClient(cc)
	if err != nil {
		log.Panic("consul客户端初始化,", err)

	}
	p.consul = client
	if err := p.register(); err != nil {
		logrus.Panic("注册服务,", err)
	}
}

func (p *app) mqttInit() {
	opts := MQTT.NewClientOptions()
	var host = viper.GetString("mqtt.host")
	var port = viper.GetInt("mqtt.port")
	var username = viper.GetString("mqtt.username")
	var password = viper.GetString("mqtt.password")
	var clientID = viper.GetString("mqtt.clientId")
	if clientID == "" {
		clientID = p.serviceId
	}
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", host, port))
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetConnectTimeout(time.Second * 20)
	opts.SetKeepAlive(time.Second * 60)
	opts.SetProtocolVersion(4)
	opts.SetClientID(clientID)
	//opts.SetConnectionLostHandler()
	opts.SetConnectionLostHandler(func(client MQTT.Client, e error) {
		if e != nil {
			logrus.Panic(e)
		}
	})
	opts.SetOrderMatters(false)
	// Start the connection
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		logrus.Panic(token.Error())
	}
	p.mqtt = client
}

func (p *app) rabbitInit() {
	if p.dataAction != "rabbit" {
		return
	}
	var (
		host     = viper.GetString("rabbit.host")
		port     = viper.GetInt("rabbit.port")
		username = viper.GetString("rabbit.username")
		password = viper.GetString("rabbit.password")
		vhost    = viper.GetString("rabbit.vhost")
	)
	var err error
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/%s", username, password, host, port, vhost))
	if err != nil {
		logrus.Panic(err)
	}
	p.rabbit = conn
}

// Register 服务注册
func (p *app) register() error {
	var (
		serviceID   = viper.GetString("service.id")
		serviceName = viper.GetString("service.name")
		serviceHost = viper.GetString("service.host")
		servicePort = viper.GetInt("service.port")
		serviceTag  = viper.GetString("service.tag")
	)

	if servicePort == 0 {
		servicePort = 9000
	}
	if serviceName == "" {
		serviceName = p.driverName
	}
	if serviceID == "" {
		logrus.Panic("服务id不能为空")
	}
	p.serviceId = serviceID
	p.serviceName = serviceName
	var err error
	serviceHost, err = extract(serviceHost)
	if err != nil {
		serviceHost = "127.0.0.1"
	}
	tags := strings.Split(serviceTag, ",")
	if len(tags) == 0 {
		tags = append(tags, "traefik.enable=false", fmt.Sprintf("driver.id=%s", p.driverId), fmt.Sprintf("driver.name=%s", p.driverName))
	} else {
		tags = append(tags, fmt.Sprintf("driver.id=%s", p.driverId), fmt.Sprintf("driver.name=%s", p.driverName))
	}
	var check = &consulApi.AgentServiceCheck{
		CheckID:                        serviceID,
		TCP:                            fmt.Sprintf("%s:%d", serviceHost, servicePort),
		Interval:                       fmt.Sprintf("%v", 10*time.Second),
		Timeout:                        fmt.Sprintf("%v", 30*time.Second),
		DeregisterCriticalServiceAfter: fmt.Sprintf("%v", getDeregisterTTL(30*time.Second)),
	}

	// register the service
	asr := &consulApi.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Tags:    tags,
		Port:    servicePort,
		Address: serviceHost,
		Check:   check,
	}
	// Specify consul connect
	asr.Connect = &consulApi.AgentServiceConnect{
		Native: true,
	}
	if err := p.consul.Agent().ServiceRegister(asr); err != nil {
		return err
	}
	return p.consul.Agent().CheckDeregister("service:" + serviceID)
}

func (p *app) getConfig() ([]byte, error) {
	res, err := resty.R().Get(fmt.Sprintf("http://%s/driver/driver/%s/config", p.traefikAddress, p.driverId))
	if err != nil {
		return nil, err
	}
	if res.StatusCode() == http.StatusOK {
		return res.Body(), nil
	}
	return nil, fmt.Errorf("请求状态:%d,响应:%s", res.StatusCode(), res.String())
}

// Start 开始服务
func (p *app) Start(driver Driver, handlers ...Handler) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	for _, handler := range handlers {
		handler.Start()
	}
	u := url.URL{Scheme: "ws", Host: p.traefikAddress, Path: "/driver/ws/" + p.driverId}
	var c *websocket.Conn
	go func() {
		//maxReconnectAttempts := 5
		var i = 4
		var timeConnect = 0
		for {
			var err error
			c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				timeConnect++
				i--
				if i <= 0 {
					logrus.Errorf("尝试重新连接WebSocket第 %d 次失败", timeConnect)
					logrus.Errorln("TCP连接已超时")
					os.Exit(1)
				}
				logrus.Errorf("尝试重新连接WebSocket第 %d 次失败", timeConnect)
				time.Sleep(time.Second * 10)
				continue
			}
			var handler = func() {
				for {
					var msg1 = new(wsRequest)
					err := c.ReadJSON(&msg1)
					if err != nil {
						logrus.Warnln("服务端关闭,", err.Error())
						return
					}

					var r result
					switch msg1.Action {
					case "reload":
						c, err := p.getConfig()
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
					err = c.WriteJSON(&wsResponse{RequestId: msg1.RequestId, Data: r})
					if err != nil {
						logrus.Warnln("写数据错误,", err.Error())
						return
					}
				}
			}
			i = 4
			timeConnect = 0
			handler()
		}
	}()
	go func() {
		c, err := p.getConfig()
		if err != nil {
			logrus.Panic("查询配置错误,", err.Error())
		}
		if err := driver.Start(p, c); err != nil {
			logrus.Panic("驱动启动错误,", err.Error())
		}
	}()
	sig := <-ch
	p.stop()
	close(ch)

	if err := driver.Stop(p); err != nil {
		logrus.Warnln("驱动停止,", err.Error())
	}
	for _, handler := range handlers {
		handler.Stop()
	}
	if err := c.Close(); err != nil {
		logrus.Warnln("关闭TCP服务器错误,", err.Error())
	}
	logrus.Debugln("关闭服务,", sig)
	os.Exit(0)
}

// Stop 服务停止
func (p *app) stop() {
	if p.mqtt != nil {
		p.mqtt.Disconnect(250)
	}
	if p.rabbit != nil && !p.rabbit.IsClosed() {
		if err := p.rabbit.Close(); err != nil {
			logrus.Errorln("关闭MQ错误,", err.Error())
		}
	}
	if err := p.consul.Agent().ServiceDeregister(p.serviceId); err != nil {
		logrus.Errorln("注销服务错误,", err.Error())
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
	b, err := json.Marshal(point)
	if err != nil {
		return err
	}
	if p.dataAction == "rabbit" {
		ch, err := p.rabbit.Channel()
		if err != nil {
			return err
		}
		defer func() {
			if err := ch.Close(); err != nil {
				logrus.Warnln("关闭管道,", err.Error())
			}
		}()
		return ch.Publish(
			"data",                            // exchange
			fmt.Sprintf("data.%s", point.Uid), // routing key
			false,                             // mandatory
			false,
			amqp.Publishing{
				DeliveryMode: amqp.Transient,
				ContentType:  "text/plain",
				Body:         b,
			})
	} else {
		if token := p.mqtt.Publish(fmt.Sprintf("data/%s", point.Uid), 0, false, string(b)); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}
	return nil
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
func (p *app) ConvertValue(tagTemp, raw interface{}) (interface{}, error) {

	tag := make(map[string]interface{})

	b, err := json.Marshal(tagTemp)
	if err != nil {
		return raw, err
	}
	err = json.Unmarshal(b, &tag)
	if err != nil {
		return raw, err
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
	case "string":
		rawString := reflect.ValueOf(raw).String()
		if strings.Contains(rawString, ".") {
			value, err = strconv.ParseFloat(rawString, 64)
			if err != nil {
				return raw, errors.New("值非数字")
			}
		} else {
			vInt64, err := strconv.ParseInt(rawString, 0, 64)
			if err != nil {
				return raw, errors.New("值非数字")
			}
			value = float64(vInt64)
		}
	default:
		return raw, errors.New("值非数字")
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

	return value, nil
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
