package sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

// DG 数据采集类
type DG struct {
	sync.Mutex
	mqttClient       MQTT.Client
	client           *consul.Client
	registryID       string
	registryName     string
	driverID         string
	driverName       string
	registryInterval time.Duration
	tags             []string
	gatewayAddress   string
	schema           string
	mqtt             *EmqttConfig
	serverPort       int
}

var ex = make(chan bool)

// ServiceConfig consul服务配置
type (
	// EmqttConfig 配置
	EmqttConfig struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	GCConfig struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	DriverConfig struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	RegistryConfig struct {
		ID   string   `json:"id"`
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	ServiceConfig struct {
		Schema  string          `json:"schema"`
		Consul  *GCConfig       `json:"consul"`
		Service *RegistryConfig `json:"service"`
		Driver  *DriverConfig   `json:"driver"`
		Gateway *GCConfig       `json:"gateway"`
		Server  *GCConfig       `json:"server"`
		Mqtt    *EmqttConfig    `json:"mqtt"`
	}
)

// NewDG 创建DG
func NewDG(service ServiceConfig) *DG {

	if service.Driver.ID == "" || service.Driver.Name == "" {
		panic("驱动id或name不能为空")
	}

	if service.Consul == nil {
		service.Consul = &GCConfig{Host: "consul", Port: 8500}
	} else {
		if service.Consul.Host == "" {
			service.Consul.Host = "consul"
		}
		if service.Consul.Port == 0 {
			service.Consul.Port = 8500
		}
	}

	if service.Service == nil {
		id := uuid.New().String()
		service.Service = &RegistryConfig{ID: id, Name: "driver_" + id, Tags: []string{}}
	} else {
		if service.Service.ID == "" {
			service.Service.ID = uuid.New().String()
		}

		if service.Service.Name == "" {
			service.Service.Name = "driver_" + service.Service.ID
		}
		if service.Service.Tags == nil {
			service.Service.Tags = make([]string, 0)
		}
	}

	if service.Gateway == nil {
		service.Gateway = &GCConfig{Host: "traefik", Port: 80}
	} else {
		if service.Gateway.Host == "" {
			service.Gateway.Host = "traefik"
		}
		if service.Gateway.Port == 0 {
			service.Gateway.Port = 80
		}
	}

	if service.Mqtt == nil {
		service.Mqtt = &EmqttConfig{Host: "mqtt", Port: 1883, Username: "admin", Password: "public"}
	} else {
		if service.Mqtt.Host == "" {
			service.Mqtt.Host = "mqtt"
		}
		if service.Mqtt.Port == 0 {
			service.Mqtt.Port = 1883
		}
		if service.Mqtt.Username == "" {
			service.Mqtt.Username = "admin"
		}
		if service.Mqtt.Password == "" {
			service.Mqtt.Password = "public"
		}
	}

	cli, err := Init(fmt.Sprintf("tcp://%s:%d", service.Mqtt.Host, service.Mqtt.Port), service.Mqtt.Username, service.Mqtt.Password)
	if err != nil {
		panic(err)
	}
	cc := consul.DefaultConfig()
	cc.Address = fmt.Sprintf("%s:%d", service.Consul.Host, service.Consul.Port)
	client, err := consul.NewClient(cc)
	if err != nil {
		panic(err)
	}

	dg := &DG{
		registryID:       service.Service.ID,
		registryName:     service.Service.Name,
		driverID:         service.Driver.ID,
		driverName:       service.Driver.Name,
		registryInterval: time.Second * 30,
		tags:             service.Service.Tags,
		gatewayAddress:   fmt.Sprintf("%s:%d", service.Gateway.Host, service.Gateway.Port),
		mqttClient:       cli,
		client:           client,
		schema:           service.Schema,
		mqtt:             service.Mqtt,
	}
	if service.Server != nil {
		dg.serverPort = service.Server.Port
	}
	if err := dg.register(); err != nil {
		panic(err)
	}
	go dg.run(ex)
	return dg
}

// Register 服务注册
func (p *DG) register() error {
	address := ":" + strconv.Itoa(p.serverPort)
	parts := strings.Split(address, ":")
	host := strings.Join(parts[:len(parts)-1], ":")
	port, _ := strconv.Atoi(parts[len(parts)-1])

	addr, err := extract(host)
	if err != nil {
		// best effort localhost
		addr = "127.0.0.1"
	}

	var check *consul.AgentServiceCheck

	deregTTL := getDeregisterTTL(p.registryInterval)
	check = &consul.AgentServiceCheck{
		TCP:                            fmt.Sprintf("%s:%d", addr, port),
		Interval:                       fmt.Sprintf("%v", p.registryInterval),
		Timeout:                        fmt.Sprintf("%v", time.Second*30),
		DeregisterCriticalServiceAfter: fmt.Sprintf("%v", deregTTL),
	}
	p.tags = append(p.tags, "traefik.enable=false", fmt.Sprintf("driver.id=%s", p.driverID), fmt.Sprintf("driver.name=%s", p.driverName))

	// register the service
	asr := &consul.AgentServiceRegistration{
		ID:      p.registryID,
		Name:    p.registryName,
		Tags:    p.tags,
		Port:    port,
		Address: addr,
		Check:   check,
	}

	// Specify consul connect
	asr.Connect = &consul.AgentServiceConnect{
		Native: true,
	}

	if err := p.client.Agent().ServiceRegister(asr); err != nil {
		return err
	}
	return p.client.Agent().CheckDeregister("service:" + p.registryID)
}

func (p *DG) run(exit chan bool) {
	t := time.NewTicker(p.registryInterval)
	for {
		select {
		case <-t.C:
			err := p.register()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"name": "运行注册consul",
				}).Errorln(err.Error())
			}
		case <-exit:
			t.Stop()
			return
		}
	}
}

func (p *DG) getConfig() ([]byte, error) {
	var url = fmt.Sprintf("http://%s/driver/driver/%s/config", p.gatewayAddress, p.driverID)
	//c := new(config)
	res, err := resty.R().Get(url)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() == http.StatusOK {
		return res.Body(), nil
	}
	return nil, fmt.Errorf("请求状态:%d,响应:%s", res.StatusCode(), res.String())
}

func (p *DG) cacheSchema(schema string) error {
	var url = fmt.Sprintf("http://%s/driver/driver/%s/schema", p.gatewayAddress, p.driverID)
	res, err := resty.R().SetBody(schema).Post(url)
	if err != nil {
		return err
	}
	if res.StatusCode() == http.StatusOK {
		return nil
	}
	return fmt.Errorf("请求状态:%d,响应:%s", res.StatusCode(), res.String())
}

// Start 开始服务
func (p *DG) Start(driver Driver) error {
	if err := p.cacheSchema(p.schema); err != nil {
		return err
	}
	// topic:command/驱动ID/请求ID
	// topic:command/驱动ID/节点ID/请求ID
	if err := Rec(p.mqttClient, fmt.Sprintf("command/%s/%s/#", p.driverID, p.registryID), func(client MQTT.Client, message MQTT.Message) {
		topics := strings.Split(message.Topic(), "/")
		logrus.WithFields(logrus.Fields{"name": "接收消息数据",}).Debugln(string(message.Payload()))
		// 重新获取配置
		if len(topics) == 4 {
			if err := driver.Reload(p, message.Payload()); err != nil {
				logrus.WithFields(logrus.Fields{"name": "修改配置,重新启动",}).Warnln(err.Error())
				b, _ := json.Marshal(map[string]interface{}{"code": http.StatusBadRequest, "message": err.Error()})
				err = Send(p.mqttClient, fmt.Sprintf("command/%s", topics[3]), b)
				if err != nil {
					logrus.WithFields(logrus.Fields{"name": "修改配置,重新启动失败.发送消息",}).Warnln(err.Error())
				}
			} else {
				b, _ := json.Marshal(map[string]interface{}{"code": http.StatusOK, "message": "驱动重启成功"})
				err = Send(p.mqttClient, fmt.Sprintf("command/%s", topics[3]), b)
				if err != nil {
					logrus.WithFields(logrus.Fields{"name": "修改配置,重新启动成功.发送消息",}).Warnln(err.Error())
				}
			}
		} else if len(topics) == 5 {
			if err := driver.Run(p, topics[3], message.Payload()); err != nil {
				logrus.WithFields(logrus.Fields{"name": "修改指令,运行错误",}).Warnln(err.Error())
				b, _ := json.Marshal(map[string]interface{}{"code": http.StatusBadRequest, "message": err.Error()})
				err = Send(p.mqttClient, fmt.Sprintf("command/%s", topics[4]), b)
				if err != nil {
					logrus.WithFields(logrus.Fields{"name": "修改指令,运行错误.发送消息",}).Warnln(err.Error())
				}
			} else {
				b, _ := json.Marshal(map[string]interface{}{"code": http.StatusOK, "message": "指令修改成功"})
				err = Send(p.mqttClient, fmt.Sprintf("command/%s", topics[4]), b)
				if err != nil {
					logrus.WithFields(logrus.Fields{"name": "修改指令,运行成功.发送消息",}).Warnln(err.Error())
				}
			}
		}
	}); err != nil {
		return err
	}
	go func() {
		for {
			c, err := p.getConfig()
			if err == nil {
				if err := driver.Start(p, c); err != nil {
					logrus.WithFields(logrus.Fields{"name": "启动错误",}).Errorln(err.Error())
				}
				break
			}
			logrus.WithFields(logrus.Fields{"name": "查询配置错误",}).Warnln(err.Error())
			time.Sleep(10 * time.Second)
		}
	}()
	return nil
}

// Stop 服务停止
func (p *DG) Stop() error {
	close(ex)
	if p.mqttClient != nil {
		p.mqttClient.Disconnect(250)
	}
	return p.client.Agent().ServiceDeregister(p.registryID)
}

// WritePoints 写数据点数据
func (p *DG) WritePoints(uid, modelID, nodeID string, fields map[string]interface{}) error {
	b, err := json.Marshal(map[string]interface{}{
		"modelId": modelID,
		"nodeId":  nodeID,
		"fields":  fields,
	})
	if err != nil {
		return err
	}
	return Send(p.mqttClient, fmt.Sprintf("data/%s", uid), b)
}

// LogDebug 写日志数据
func (p *DG) Log(topic string, msg interface{}) {
	l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "message": msg}
	b, err := json.Marshal(l)
	if err != nil {
		return
	}
	Send(p.mqttClient, "logs/"+topic, b)
}

// LogDebug 写日志数据
func (p *DG) LogDebug(uid string, msg interface{}) {
	l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "uid": uid, "message": msg}
	b, err := json.Marshal(l)
	if err != nil {
		return
	}
	Send(p.mqttClient, "logs/debug/"+uid, b)
}

// LogInfo 写日志数据
func (p *DG) LogInfo(uid string, msg interface{}) {
	l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "uid": uid, "message": msg}
	b, err := json.Marshal(l)
	if err != nil {
		return
	}
	Send(p.mqttClient, "logs/info/"+uid, b)
}

// LogWarn 写日志数据
func (p *DG) LogWarn(uid string, msg interface{}) {
	l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "uid": uid, "message": msg}
	b, err := json.Marshal(l)
	if err != nil {
		return
	}
	Send(p.mqttClient, "logs/warn/"+uid, b)
}

// LogError 写日志数据
func (p *DG) LogError(uid string, msg interface{}) {
	l := map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "uid": uid, "message": msg}
	b, err := json.Marshal(l)
	if err != nil {
		return
	}
	Send(p.mqttClient, "logs/error/"+uid, b)
}

// ConvertValue 数据点值转换
func (p *DG) ConvertValue(tagTemp, raw interface{}) (interface{}, error) {

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
	var valueRaw float64
	vType := reflect.TypeOf(raw).String()

	switch vType {
	case "float32", "float64":
		value = reflect.ValueOf(raw).Float()
		valueRaw = reflect.ValueOf(raw).Float()
	case "uint", "uintptr", "uint8", "uint16", "uint32", "uint64":
		value = float64(reflect.ValueOf(raw).Uint())
		valueRaw = float64(reflect.ValueOf(raw).Uint())
	case "int", "int8", "int16", "int32", "int64":
		value = float64(reflect.ValueOf(raw).Int())
		valueRaw = float64(reflect.ValueOf(raw).Int())
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
		value = (((valueRaw - minRaw) / (maxRaw - minRaw)) * (maxValue - minValue)) + minValue
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

func (p *DG) convertValue(tagValue *map[string]float64, key string, tagValueMap map[string]interface{}) {
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
