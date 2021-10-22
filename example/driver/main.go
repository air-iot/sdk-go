package main

import (
	"encoding/json"
	"github.com/air-iot/sdk-go/driver"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

type (
	// 驱动配置信息，不同的驱动生成不同的配置信息
	config []model

	// model 模型信息
	model struct {
		ID      string `json:"id"`     // 模型id，模型唯一标识
		Device  Device `json:"device"` // 模型驱动信息
		Devices []struct {
			ID     string `json:"id"`     // 资产id，资产唯一标识
			Uid    string `json:"uid"`    // 资产的唯一编号
			Device Device `json:"device"` // 资产驱动信息
		} `json:"devices"` // 所属模型的资产配置信息
	}

	// 驱动信息
	Device struct {
		Driver string `json:"driver"` // 驱动名称
		Tags   []struct {
			ID   string `json:"id"`   // 数据点唯一标识
			Name string `json:"name"` // 数据点名称
		} `json:"tags"` // 驱动数据点
		Commands []struct {
			ID   string `json:"id"`   // 指令唯一标识
			Name string `json:"name"` // 指令名称
		} `json:"commands"` // 指令配置
	}
)

// TestDriver 定义测试驱动结构体
type TestDriver struct{}

// Start 驱动执行，实现Driver的Start函数
func (p *TestDriver) Start(a driver.App, models []byte) error {
	a.GetLogger().Debugln("start", string(models))
	ms := config{}
	err := json.Unmarshal(models, &ms)
	if err != nil {
		a.GetLogger().Errorln(err)
		return err
	}
	for _, m1 := range ms {
		if m1.Devices == nil {
			continue
		}
		for _, n1 := range m1.Devices {
			if n1.Device.Tags == nil {
				continue
			}
			fields := make([]driver.Field, 0)
			fieldType := make(map[string]string)
			if m1.Device.Tags != nil {
				for _, t1 := range m1.Device.Tags {
					// fields[t1.ID] = rand.Intn(100)
					fields = append(fields, driver.Field{Tag: t1, Value: rand.Intn(100)})
					fieldType[t1.ID] = driver.Integer
				}
			}
			for _, t1 := range n1.Device.Tags {
				// fields[t1.ID] = rand.Intn(100)
				fields = append(fields, driver.Field{Tag: t1, Value: rand.Intn(100)})

			}
			for tries := 0; tries < 1000; tries++ {
				point := driver.Point{
					ID:         n1.ID,
					Fields:     fields,
					FieldTypes: fieldType,
					UnixTime:   time.Now().UnixNano() / 1e6,
				}
				if err := a.WritePoints(point); err != nil {
					// a.LogError(n1.Uid, "写数据错误")
					a.GetLogger().Errorln("写数据,", err)
				}
				logrus.Debugf("写数据成功,%+v", point)
				time.Sleep(time.Second)
			}
		}
	}
	return nil
}

func (p *TestDriver) Schema(a driver.App) (string, error) {
	return "测试", nil
}

// Reload 驱动重启，实现Driver的Reload函数
func (p *TestDriver) Reload(a driver.App, models []byte) error {
	a.GetLogger().Debugln("Reload")
	return nil
}

// Run 执行指令，实现Driver的Run函数
func (p *TestDriver) Run(a driver.App, deviceID string, cmd []byte) (interface{}, error) {
	a.GetLogger().Debugln("run", deviceID, string(cmd))
	return nil, nil
}

func (p *TestDriver) WriteTag(a driver.App, deviceID string, cmd []byte) (interface{}, error) {
	a.GetLogger().Debugln("WriteTag", deviceID, string(cmd))
	return nil, nil
}

func (p *TestDriver) Debug(a driver.App, b []byte) (interface{}, error) {
	a.GetLogger().Debugln("debug", string(b))
	return []int{1, 2, 3}, nil
}

func (p *TestDriver) Stop(a driver.App) error {
	a.GetLogger().Debugln("stop")
	return nil
}

func main() {
	// 创建采集主程序
	driver.NewApp().Start(new(TestDriver))
}
