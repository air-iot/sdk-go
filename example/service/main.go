package main

import (
	"github.com/air-iot/sdk-go/service"
	"github.com/labstack/echo/v4"
	"net/http"
)

// TestService 定义测试接口结构体
type TestService struct {
}

// Start 驱动执行，实现Driver的Start函数
func (p *TestService) Start(a service.App) error {
	a.GetLogger().Debugln("start")
	a.GetHttpServer().GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world")
	})
	return nil
}

func (p *TestService) Stop(a service.App) error {
	a.GetLogger().Debugln("stop")
	return nil
}

func main() {
	// 创建接口主程序
	service.NewApp().Start(new(TestService))
}
