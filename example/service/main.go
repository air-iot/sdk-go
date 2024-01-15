/*
	* @Descripttion:
	* @version:
	* @Author: zhangqiang
	* @Date: 2020-08-06 14:12:01
  * @LastEditors: zhangqiang
  * @LastEditTime: 2020-08-07 11:27:59
*/
package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"

	"github.com/air-iot/sdk-go/v4/service"
)

// TestService 定义测试接口结构体
type TestService struct {
}

// Start 驱动执行，实现Service的Start函数
func (p *TestService) Start(a service.App) error {
	log.Println("start")
	a.GetHttpServer().GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello world")
	})
	return nil
}

// Stop 接口停止，实现Service的Stop函数
func (p *TestService) Stop(a service.App) error {
	log.Println("stop")
	return nil
}

func main() {
	// 创建接口主程序
	service.NewApp().Start(new(TestService))
}
