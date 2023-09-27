package app

import (
	"context"
	"github.com/air-iot/logger"
	"github.com/air-iot/sdk-go/v4/flow"
)

// TestFlow 定义测试驱动结构体
type TestFlow struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

func (p *TestFlow) Handler(ctx context.Context, app flow.App, request *flow.Request) (map[string]interface{}, error) {
	logger.Infof("配置: %+v", *request)
	return map[string]interface{}{"a": 1}, nil
}
