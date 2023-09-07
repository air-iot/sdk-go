package app

import (
	"fmt"
	"github.com/air-iot/json"
	"github.com/air-iot/logger"
	"github.com/air-iot/sdk-go/v4/driver"
	"github.com/air-iot/sdk-go/v4/driver/entity"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"net/http"
)

// 驱动配置信息，不同的驱动生成不同的配置信息
type (
	DriverInstanceConfig struct {
		ID         string  `json:"id"`
		Name       string  `json:"name"`
		DriverType string  `json:"driverType"`
		Device     Device  `json:"device"`
		Tables     []table `json:"tables"`
	}

	table struct {
		ID      string `json:"id"`     // 模型id，模型唯一标识
		Device  Device `json:"device"` // 模型驱动信息
		Devices []struct {
			ID     string `json:"id"`     // 资产id，资产唯一标识
			Device Device `json:"device"` // 资产驱动信息
		} `json:"devices"` // 所属模型的资产配置信息
	}

	Device struct {
		Driver   string `json:"driver"` // 驱动名称
		Settings struct {
			Server        string `json:"server"`
			Username      string `json:"username"`
			Password      string `json:"password"`
			ClientId      string `json:"clientId"`
			Topic         string `json:"topic"`
			ParseScript   string `json:"parseScript"`
			CommandScript string `json:"commandScript"`
		} `json:"settings"`
		Tags     []tag `json:"tags"` // 驱动数据点
		Commands []struct {
			ID   string `json:"id"`   // 指令唯一标识
			Name string `json:"name"` // 指令名称
		} `json:"commands"` // 指令配置
	}

	tag struct {
		ID   string `json:"id"`   // 数据点唯一标识
		Name string `json:"name"` // 数据点名称
	}
)

// TestDriver 定义测试驱动结构体
type TestDriver struct {
	client         MQTT.Client
	parseVm        *goja.Runtime
	parseHandler   goja.Callable
	commandVm      *goja.Runtime
	commandHandler goja.Callable
	tables         map[string]map[string]map[string]tag
}

type parseResult struct {
	Table  string                 `json:"table"`
	Id     string                 `json:"id"`
	Time   int64                  `json:"time"`
	Fields map[string]interface{} `json:"fields"`
}

type cmdResult struct {
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
}

// Start 驱动执行，实现Driver的Start函数
func (p *TestDriver) Start(a driver.App, bts []byte) error {
	logger.Debugln("start", string(bts))
	if err := p.Stop(a); err != nil {
		return err
	}
	var config DriverInstanceConfig
	err := json.Unmarshal(bts, &config)
	if err != nil {
		return err
	}
	if config.Device.Settings.Server == "" {
		return fmt.Errorf("服务器地址为空")
	}
	if config.Device.Settings.Topic == "" {
		return fmt.Errorf("Topic为空")
	}
	registry := require.NewRegistry()
	if config.Device.Settings.ParseScript != "" {
		vm := goja.New()
		registry.Enable(vm)
		console.Enable(vm)
		if _, err := vm.RunString(config.Device.Settings.ParseScript); err != nil {
			return err
		}
		handler, ok := goja.AssertFunction(vm.Get("handler"))
		if !ok {
			return fmt.Errorf("解析脚本函数handler未找到")
		}
		p.parseVm = vm
		p.parseHandler = handler
	}
	if config.Device.Settings.CommandScript != "" {
		vm := goja.New()
		registry.Enable(vm)
		console.Enable(vm)
		if _, err := vm.RunString(config.Device.Settings.CommandScript); err != nil {
			return err
		}
		handler, ok := goja.AssertFunction(vm.Get("handler"))
		if !ok {
			return fmt.Errorf("指令脚本函数handler未找到")
		}
		p.commandVm = vm
		p.commandHandler = handler
	}
	opts := MQTT.NewClientOptions()
	opts.AddBroker(config.Device.Settings.Server)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)
	opts.SetUsername(config.Device.Settings.Username)
	opts.SetPassword(config.Device.Settings.Password)
	if config.Device.Settings.ClientId != "" {
		opts.SetClientID(config.Device.Settings.ClientId)
	}
	opts.SetConnectionLostHandler(func(client MQTT.Client, e error) {
		panic(fmt.Errorf("MQTT Lost错误: %s", e.Error()))
	})
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	p.client = client
	if err := p.handler(a, config); err != nil {
		return err
	}
	return nil
}

func (p *TestDriver) Schema(_ driver.App) (string, error) {
	return Schema, nil
}

// Run 执行指令，实现Driver的Run函数
func (p *TestDriver) Run(_ driver.App, cmd *entity.Command) (interface{}, error) {
	logger.Debugln("执行指令", *cmd)
	var c map[string]interface{}
	err := json.Unmarshal(cmd.Command, &c)
	if err != nil {
		return nil, err
	}
	result, err := p.commandHandler(goja.Undefined(), p.commandVm.ToValue(cmd.Table), p.commandVm.ToValue(cmd.Id), p.commandVm.ToValue(c))
	if err != nil {
		return nil, err
	}
	var ret cmdResult
	if err := json.CopyByJson(&ret, result.Export()); err != nil {
		return nil, err
	}
	if token := p.client.Publish(ret.Topic, 0, false, ret.Payload); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return nil, nil
}

// BatchRun 批量执行指令，实现Driver的Run函数
func (p *TestDriver) BatchRun(_ driver.App, cmd *entity.BatchCommand) (interface{}, error) {
	logger.Debugln("批量执行指令", *cmd)
	return nil, nil
}

func (p *TestDriver) WriteTag(_ driver.App, cmd *entity.Command) (interface{}, error) {
	logger.Debugln("写数据点", *cmd)
	return nil, nil
}

func (p *TestDriver) Debug(_ driver.App, b []byte) (interface{}, error) {
	logger.Debugln("调试", string(b))
	return []int{}, nil
}

func (p *TestDriver) Stop(_ driver.App) error {
	logger.Debugln("停止")
	p.parseVm = nil
	p.parseHandler = nil
	p.commandVm = nil
	p.commandHandler = nil
	if p.client != nil {
		p.client.Disconnect(250)
	}
	p.tables = map[string]map[string]map[string]tag{}
	return nil
}

func (p *TestDriver) HttpProxy(_ driver.App, t string, header http.Header, data []byte) (interface{}, error) {
	logger.Debugln("Http代理", t, header, string(data))
	return Schema, nil
}

func (p *TestDriver) handler(a driver.App, driverConfig DriverInstanceConfig) error {
	for _, t := range driverConfig.Tables {
		if len(t.Devices) == 0 {
			continue
		}
		tagMap := map[string]tag{}
		for _, ta := range t.Device.Tags {
			tagMap[ta.ID] = ta
		}
		dev1 := map[string]map[string]tag{}
		p.tables[t.ID] = dev1
		for _, device := range t.Devices {
			devTagMap := map[string]tag{}
			for k, v := range tagMap {
				devTagMap[k] = v
			}
			for _, tagE := range device.Device.Tags {
				devTagMap[tagE.ID] = tagE
			}
			dev1[device.ID] = devTagMap
		}
	}
	p.client.Subscribe(driverConfig.Device.Settings.Topic, 0, func(client MQTT.Client, message MQTT.Message) {
		if p.parseHandler == nil || p.parseVm == nil {
			logger.Errorln("解析脚本为空")
			return
		}
		var data []map[string]interface{}
		err := json.Unmarshal(message.Payload(), &data)
		if err != nil {
			logger.Errorf("消息解析,%v", err)
			return
		}
		result, err := p.parseHandler(goja.Undefined(), p.parseVm.ToValue(message.Topic()), p.parseVm.ToValue(data))
		if err != nil {
			logger.Errorf("消息解析错误,%v", err)
			return
		}

		var arr []parseResult
		if err := json.CopyByJson(&arr, result.Export()); err != nil {
			logger.Errorf("实例执行脚本结果解序列化错误,%s", err)
			return
		}
		for _, v := range arr {
			dev, ok := p.tables[v.Table]
			if !ok {
				continue
			}
			tagM, ok := dev[v.Id]
			if !ok {
				continue
			}
			fields := make([]entity.Field, 0)
			for k1, v1 := range v.Fields {
				tagT, ok := tagM[k1]
				if !ok {
					continue
				}
				fields = append(fields, entity.Field{
					Tag:   tagT,
					Value: v1,
				})
			}
			err := a.WritePoints(entity.Point{
				Table:    v.Table,
				ID:       v.Id,
				Fields:   fields,
				UnixTime: v.Time,
			})
			if err != nil {
				logger.Errorf("写数据错误,%v", err)
				return
			}
		}
	})
	return nil
}
