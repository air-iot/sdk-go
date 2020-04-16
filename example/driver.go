package main

import (
	"encoding/json"
	"log"
	"math/rand"

	"github.com/air-iot/sdk"
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
func (p *TestDriver) Start(a sdk.App, models []byte) error {
	log.Println("start", string(models))
	ms := config{}
	err := json.Unmarshal(models, &ms)
	if err != nil {
		log.Println(err)
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
			fields := make(map[string]interface{})
			if m1.Device.Tags != nil {
				for _, t1 := range m1.Device.Tags {
					fields[t1.ID] = rand.Intn(100)
				}
			}
			for _, t1 := range n1.Device.Tags {
				fields[t1.ID] = rand.Intn(100)
			}
			point := sdk.Point{
				Uid:      n1.Uid,
				ModelId:  m1.ID,
				NodeId:   n1.ID,
				Fields:   fields,
				UnixTime: 0,
			}
			if err := a.WritePoints(point); err != nil {
				a.LogError(n1.Uid, "写数据错误")
			}

		}
	}
	return nil
}

// Reload 驱动重启，实现Driver的Reload函数
func (p *TestDriver) Reload(a sdk.App, models []byte) error {
	return p.Start(a, models)
}

// Run 执行指令，实现Driver的Run函数
func (p *TestDriver) Run(a sdk.App, deviceID string, cmd []byte) error {
	log.Println("run", deviceID, string(cmd))
	return nil
}

func (p *TestDriver) Debug(a sdk.App, b []byte) (interface{}, error) {
	log.Println("debug", string(b))
	return []int{1, 2, 3}, nil
}

func main() {
	// 创建采集主程序
	sdk.NewApp().Start(new(TestDriver))
}
