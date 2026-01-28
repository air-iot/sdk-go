package driver

import (
	"compress/gzip"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/air-iot/logger"
	"github.com/gin-gonic/gin"
)

// recoveryMiddleware 恢复中间件，捕获 panic 并记录日志
func recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := make([]byte, 4096)
				length := runtime.Stack(stack, false)
				logger.Errorf("[PANIC] %v\n堆栈信息:\n%s", err, stack[:length])
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "服务器内部错误",
					"code":  500,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// loggerMiddleware 日志中间件，记录请求信息
func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 记录请求日志
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if query != "" {
			path = path + "?" + query
		}

		logger.Infof("[HTTP] %s %s | 状态=%d | 耗时=%v | IP=%s",
			method,
			path,
			status,
			latency,
			clientIP,
		)

		// 记录错误状态
		if status >= 400 {
			logger.Warnf("[HTTP] 请求失败: %s %s | 状态=%d", method, path, status)
		}
	}
}

// corsMiddleware CORS 跨域中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// 处理 OPTIONS 预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// gzipMiddleware Gzip 压缩中间件
func gzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查客户端是否支持 gzip
		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// 检查响应类型，不对已压缩的内容再次压缩
		if c.Writer.Header().Get("Content-Encoding") != "" {
			c.Next()
			return
		}

		// 跳过静态资源（图片、字体、视频等）
		// Gin Static 会设置 Content-Type，我们在中间件中检查
		// 但此时还没设置，所以需要在 Write 时动态判断

		// 使用自定义的 ResponseWriter
		gz := &gzipWriter{
			ResponseWriter: c.Writer,
			gzipWriter:     nil, // 延迟创建
			gzipped:        false,
		}
		c.Writer = gz
		c.Next()

		// 如果实际使用了 gzip，设置响应头
		if gz.gzipped && gz.gzipWriter != nil {
			gz.gzipWriter.Close()
		}
	}
}

// gzipWriter 自定义 ResponseWriter，包装 gzip.Writer
type gzipWriter struct {
	gin.ResponseWriter
	gzipWriter *gzip.Writer
	gzipped    bool
}

// Write 写入数据，自动压缩
func (g *gzipWriter) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	// 检查内容类型，只压缩文本内容
	contentType := g.Header().Get("Content-Type")
	shouldCompress := strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "text/") ||
		strings.HasPrefix(contentType, "application/xml") ||
		strings.HasPrefix(contentType, "application/javascript")

	if shouldCompress {
		// 延迟创建 gzip writer
		if g.gzipWriter == nil {
			g.gzipWriter = gzip.NewWriter(g.ResponseWriter)
			g.gzipped = true
			// 只在实际压缩时设置响应头
			g.Header().Set("Content-Encoding", "gzip")
			g.Header().Set("Vary", "Accept-Encoding")
		}
		return g.gzipWriter.Write(data)
	}
	return g.ResponseWriter.Write(data)
}

// WriteString 写入字符串
func (g *gzipWriter) WriteString(s string) (int, error) {
	if len(s) == 0 {
		return 0, nil
	}
	return g.Write([]byte(s))
}
