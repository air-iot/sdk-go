package app

import (
	"context"
	"fmt"
	"github.com/air-iot/json"
	"math/rand"
	"time"

	"github.com/air-iot/logger"
	"github.com/air-iot/sdk-go/v4/driver"
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
type TestDriver struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

// Start 驱动执行，实现Driver的Start函数
func (p *TestDriver) Start(a driver.App, bts []byte) error {
	logger.Debugln("start", string(bts))
	var config DriverInstanceConfig
	err := json.Unmarshal(bts, &config)
	if err != nil {
		return err
	}
	p.Cancel()
	time.Sleep(time.Second * 2)
	p.Ctx, p.Cancel = context.WithCancel(context.Background())
	go func(config DriverInstanceConfig) {
		for {
			select {
			case <-p.Ctx.Done():
				logger.Infof("测试驱动停止")
				return
			default:
				for _, t1 := range config.Tables {
					if t1.Devices == nil {
						continue
					}
					for _, n1 := range t1.Devices {
						//if n1.Device.Tags == nil {
						//	continue
						//}
						fields := make([]driver.Field, 0)
						fieldType := make(map[string]string)
						if t1.Device.Tags != nil {
							for _, t1 := range t1.Device.Tags {
								// fields[t1.ID] = rand.Intn(100)
								fields = append(fields, driver.Field{Tag: t1, Value: rand.Intn(100)})
								fieldType[t1.ID] = driver.Integer
							}
						}
						if n1.Device.Tags != nil {
							for _, t1 := range n1.Device.Tags {
								// fields[t1.ID] = rand.Intn(100)
								fields = append(fields, driver.Field{Tag: t1, Value: rand.Intn(100)})
							}
						}
						point := driver.Point{
							//Table:      t1.ID,
							ID:         n1.ID,
							Fields:     fields,
							FieldTypes: fieldType,
							UnixTime:   time.Now().UnixNano() / 1e6,
						}
						if err := a.WritePoints(point); err != nil {
							// a.LogError(n1.Uid, "写数据错误")
							logger.Errorln("写数据,", err)
						}
					}
				}
			}
			time.Sleep(time.Second * 1)
		}
	}(config)

	return nil
}

func (p *TestDriver) Schema(a driver.App) (string, error) {
	return Schema, nil
}

// Run 执行指令，实现Driver的Run函数
func (p *TestDriver) Run(a driver.App, cmd *driver.Command) (interface{}, error) {
	fmt.Println("run", *cmd)
	logger.Debugln("run", *cmd)
	//if err := a.RunLog(driver.Log{
	//	SerialNo: cmd.SerialNo,
	//	Status:   "成功",
	//	UnixTime: time.Now().UnixNano() / 1e6,
	//	Desc:     "测试",
	//}); err != nil {
	//	logger.Errorf("run log: %s", err)
	//}
	//
	//if err := a.WriteEvent(driver.Event{
	//	Table:    cmd.Table,
	//	ID:       cmd.Id,
	//	EventID:  "test11",
	//	UnixTime: time.Now().UnixNano() / 1e6,
	//	Data:     cmd.Command,
	//}); err != nil {
	//	logger.Errorf("writeEvent: %s", err)
	//}

	//if err := a.UpdateNode(cmd.NodeId, map[string]interface{}{"n2": 22}); err != nil {
	//	a.GetLogger().Errorf("UpdateNode: %s", err)
	//}
	return nil, nil
}

// BatchRun 批量执行指令，实现Driver的Run函数
func (p *TestDriver) BatchRun(a driver.App, cmd *driver.BatchCommand) (interface{}, error) {
	logger.Debugln("BatchRun", *cmd)
	if err := a.RunLog(context.Background(), driver.Log{
		SerialNo: cmd.SerialNo,
		Status:   "成功",
		UnixTime: time.Now().UnixNano() / 1e6,
		Desc:     "批量测试",
	}); err != nil {
		logger.Errorf("run log: %s", err)
	}
	return nil, nil
}

func (p *TestDriver) WriteTag(a driver.App, cmd *driver.Command) (interface{}, error) {
	logger.Debugln("WriteTag", *cmd)
	return nil, nil
}

func (p *TestDriver) Debug(a driver.App, b []byte) (interface{}, error) {
	logger.Debugln("debug", string(b))
	return []int{1, 2, 3}, nil
}

func (p *TestDriver) Stop(a driver.App) error {
	logger.Debugln("stop")
	p.Cancel()
	return nil
}
