package main

import (
	"context"

	"github.com/air-iot/sdk-go/v4/example/flow_extension/app"
	flowextionsion "github.com/air-iot/sdk-go/v4/flow_extension"
)

func main() {
	// 创建采集主程序
	d := new(app.TestFlow)
	d.Ctx, d.Cancel = context.WithCancel(context.Background())
	flowextionsion.NewApp().Start(d)
}
