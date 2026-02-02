package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/air-iot/logger"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/air-iot/sdk-go/v4/driver/entity"
)

// initRouter 初始化 gin Router
func (a *app) initRouter() {
	gin.SetMode(gin.ReleaseMode)
	a.router = gin.New()

	// 自定义中间件
	a.router.Use(recoveryMiddleware())
	a.router.Use(loggerMiddleware())
	a.router.Use(corsMiddleware())
	a.router.Use(gzipMiddleware())

	// 注册 HTML 静态文件服务
	htmlDir := "."
	if _, err := os.Stat(htmlDir); err == nil {
		// 服务 bundle 静态资源（CSS 和 JS）
		a.router.GET("/iot-config.bundle.js", func(c *gin.Context) {
			c.File(filepath.Join(htmlDir, "iot-config.bundle.js"))
			c.Header("Content-Type", "text/javascript")
		})
		// 服务 CSS 文件
		a.router.GET("/iot-config.bundle.css", func(c *gin.Context) {
			c.File(filepath.Join(htmlDir, "iot-config.bundle.css"))
			c.Header("Content-Type", "text/css")
		})
		// 根路径返回 index.html
		a.router.GET("/", func(c *gin.Context) {
			c.File(filepath.Join(htmlDir, "index.html"))
		})
		logger.Infof("HTML静态文件服务已启用: %s", htmlDir)
	}

	// 注册默认路由
	a.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": Cfg.Driver.Name,
		})
	})

	a.router.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"serviceId": Cfg.ServiceID,
			"projectId": Cfg.Project,
			"groupId":   Cfg.GroupID,
			"driver":    Cfg.Driver,
		})
	})

	// 如果启用 pprof，注册 pprof 路由
	if Cfg.Pprof.Enable {
		pprof.Register(a.router)
		logger.Infof("pprof路由已注册到HTTP服务器: /debug/pprof/*")
	}

	// 注册 driver 方法调用接口
	api := a.router.Group("/driver")
	{
		// Schema 获取驱动配置 Schema
		api.GET("/schema", func(c *gin.Context) {
			if a.driver == nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "驱动服务尚未初始化，请稍后重试"})
				return
			}
			// 优先从请求头获取 locale，其次使用查询参数，默认 "zh"
			locale := c.GetHeader("locale")
			if locale == "" {
				locale = c.GetHeader("Accept-Language")
			}
			if locale == "" {
				locale = c.DefaultQuery("locale", "zh")
			}
			schema, err := a.driver.Schema(c.Request.Context(), a, locale)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取驱动配置Schema失败: %v", err)})
				return
			}
			c.JSON(http.StatusOK, gin.H{"schema": schema})
		})

		// GetConfig 获取当前驱动配置
		api.GET("/config", func(c *gin.Context) {
			// 从 data.json 读取当前配置
			configData, err := os.ReadFile(Cfg.DataConfig)
			if err != nil {
				if os.IsNotExist(err) {
					// 文件不存在时返回空的驱动配置
					emptyConfig := map[string]interface{}{
						"id":         Cfg.ServiceID,
						"name":       Cfg.Driver.Name,
						"groupId":    Cfg.Project,
						"driverType": Cfg.Driver.ID,
						"runMode":    "one",
						"device": map[string]interface{}{
							"commands": []interface{}{},
							"settings": map[string]interface{}{},
						},
						"distributed": "all",
						"ports":       "",
						"tables":      []interface{}{},
					}
					b, _ := json.Marshal(emptyConfig)
					c.Data(http.StatusOK, "application/json", b)
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("读取配置文件失败: %v", err)})
				return
			}

			// 直接返回 data.json 的内容（驱动格式）
			c.Header("Content-Type", "application/json")
			c.Data(http.StatusOK, "application/json", configData)
		})

		// Start 启动驱动
		api.POST("/start", func(c *gin.Context) {
			if a.driver == nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "驱动服务尚未初始化，请稍后重试"})
				return
			}

			// 获取请求数据（前端发送的已经是驱动格式）
			driverConfig, err := c.GetRawData()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("请求数据格式错误: %v", err)})
				return
			}

			// 保存配置到 data.json
			if err := a.saveDataConfig(driverConfig); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("保存配置错误: %v", err)})
				return
			}

			// 调用驱动 Start 方法
			err = a.driver.Start(c.Request.Context(), a, driverConfig)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("启动驱动失败: %v", err)})
				return
			}

			c.JSON(http.StatusOK, gin.H{"result": "启动成功"})
		})

		// Run 运行指令
		api.POST("/run", func(c *gin.Context) {
			if a.driver == nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "驱动服务尚未初始化，请稍后重试"})
				return
			}
			var command entity.RequestCommand
			if err := c.ShouldBindJSON(&command); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("请求参数格式错误: %v，请检查 JSON 格式是否正确", err)})
				return
			}
			cmdBs, _ := json.Marshal(command)
			cmd := entity.Command{
				Table:    command.Table,
				Id:       command.Id,
				SerialNo: command.Table + command.Id,
				Command:  cmdBs,
			}

			result, err := a.driver.Run(c.Request.Context(), a, &cmd)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("执行指令失败: %v", err)})
				return
			}
			c.JSON(http.StatusOK, gin.H{"result": result})
		})

		// BatchRun 批量运行指令
		api.POST("/batch-run", func(c *gin.Context) {
			if a.driver == nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "驱动服务尚未初始化，请稍后重试"})
				return
			}
			var command entity.BatchCommand
			if err := c.ShouldBindJSON(&command); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("请求参数格式错误: %v，请检查 JSON 格式是否正确", err)})
				return
			}
			result, err := a.driver.BatchRun(c.Request.Context(), a, &command)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("批量执行指令失败: %v", err)})
				return
			}
			c.JSON(http.StatusOK, gin.H{"result": result})
		})

		// WriteTag 数据点写入
		api.POST("/write-tag", func(c *gin.Context) {
			if a.driver == nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "驱动服务尚未初始化，请稍后重试"})
				return
			}
			var command entity.Command
			if err := c.ShouldBindJSON(&command); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("请求参数格式错误: %v，请检查 JSON 格式是否正确", err)})
				return
			}
			result, err := a.driver.WriteTag(c.Request.Context(), a, &command)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("写入数据点失败: %v", err)})
				return
			}
			c.JSON(http.StatusOK, gin.H{"result": result})
		})

		// Debug 调试驱动
		api.POST("/debug", func(c *gin.Context) {
			if a.driver == nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "驱动服务尚未初始化，请稍后重试"})
				return
			}
			debugConfig, err := c.GetRawData()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("请求数据格式错误: %v", err)})
				return
			}
			result, err := a.driver.Debug(c.Request.Context(), a, debugConfig)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("调试驱动失败: %v", err)})
				return
			}
			c.JSON(http.StatusOK, gin.H{"result": result})
		})

		// Realtime WebSocket 实时数据推送
		api.GET("/ws", func(c *gin.Context) {
			a.handleRealtimeWebSocket(c)
		})
	}

	// SPA 路由支持：将所有未匹配的路由重定向到 index.html
	// 这必须在所有其他路由之后注册
	if _, err := os.Stat(htmlDir); err == nil {
		a.router.NoRoute(func(c *gin.Context) {
			requestedPath := c.Request.URL.Path

			// 如果是 API 请求，返回 404
			if strings.HasPrefix(requestedPath, "/driver") || strings.HasPrefix(requestedPath, "/debug") {
				c.JSON(http.StatusNotFound, gin.H{"error": "接口不存在"})
				return
			}

			// 对于非根路径，先尝试提供静态文件
			if requestedPath != "/" && requestedPath != "" {
				// 移除开头的 /，确保路径正确拼接
				cleanPath := strings.TrimPrefix(requestedPath, "/")
				filePath := filepath.Join(htmlDir, cleanPath)
				// 检查文件是否存在且是文件（不是目录）
				if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
					c.File(filePath)
					return
				}
			}

			// 其他请求返回 index.html，让 SPA 处理路由
			c.File(filepath.Join(htmlDir, "index.html"))
		})
	}
}

// GetRouter 获取以 driver.id 为路径前缀的 RouterGroup
func (a *app) GetRouter() *gin.RouterGroup {
	if a.router == nil {
		return nil
	}
	return a.router.Group("/" + Cfg.Driver.ID)
}

// StartHTTPServer 启动 HTTP 服务（在注册路由后调用）
func (a *app) StartHTTPServer() error {
	if a.router == nil {
		return fmt.Errorf("HTTP 路由未初始化，请检查配置文件中 http.host 和 http.port 是否正确配置")
	}
	if a.httpServer != nil {
		return fmt.Errorf("HTTP 服务已在运行中，无需重复启动")
	}

	addr := net.JoinHostPort(Cfg.HTTP.Host, Cfg.HTTP.Port)
	a.httpServer = &http.Server{
		Addr:    addr,
		Handler: a.router,
	}

	go func() {
		logger.Infof("HTTP服务启动: 地址=%s", addr)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("HTTP服务启动失败: %v", err)
		}
	}()

	a.httpClean = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.httpServer.Shutdown(ctx); err != nil {
			logger.Errorf("HTTP服务关闭失败: %v", err)
		} else {
			logger.Infof("HTTP服务已关闭")
		}
	}

	return nil
}

// handleRealtimeWebSocket 处理实时数据 WebSocket 连接
func (a *app) handleRealtimeWebSocket(c *gin.Context) {
	// WebSocket 升级器配置
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// 允许所有来源，生产环境应根据需要限制
			return true
		},
	}

	// 获取连接参数
	typ := c.DefaultQuery("type", "data") // data, tag, device, model
	table := c.DefaultQuery("table", "")
	device := c.DefaultQuery("device", "")

	// 升级 HTTP 连接到 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Errorf("WebSocket升级失败: %v", err)
		return
	}

	// 创建客户端连接
	wsConn := &websocketConn{
		conn:   conn,
		typ:    typ,
		table:  table,
		device: device,
	}

	// 注册客户端
	a.registerWebSocketClient(wsConn)

	logger.Infof("WebSocket客户端已连接: type=%s, table=%s, device=%s", typ, table, device)

	// 启动读取协程（保持连接活跃，处理心跳）
	go wsConn.readPump(a)
}

// readPump 从 WebSocket 读取消息（用于保持连接活跃、处理订阅消息和客户端断开）
func (ws *websocketConn) readPump(a *app) {
	conn := ws.conn.(*websocket.Conn)
	defer func() {
		a.unregisterWebSocketClient(ws)
		_ = conn.Close()
	}()

	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	// 配置 pong 消息处理
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Errorf("WebSocket读取错误: %v", err)
			}
			break
		}

		// 处理客户端发送的消息（订阅/取消订阅）
		ws.handleMessage(message)
	}
}

// handleMessage 处理客户端发送的消息
func (ws *websocketConn) handleMessage(data []byte) {
	var msg struct {
		Table  string `json:"table"`  // 模型ID
		Device string `json:"device"` // 设备ID
	}

	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Errorf("WebSocket消息解析失败: %v", err)
		return
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.table = msg.Table
	ws.device = msg.Device
	logger.Infof("WebSocket客户端订阅: type=%s, table=%s, device=%s", ws.typ, msg.Table, msg.Device)
}
