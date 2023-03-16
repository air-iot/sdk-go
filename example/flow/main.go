package main

import (
	"context"
	"github.com/air-iot/sdk-go/v4/example/flow/app"
	"github.com/air-iot/sdk-go/v4/flow"
)

func main() {
	// 创建采集主程序
	d := new(app.TestFlow)
	d.Ctx, d.Cancel = context.WithCancel(context.Background())
	flow.NewApp().Start(d)
}
