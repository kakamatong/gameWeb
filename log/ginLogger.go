package log

import (
	"time"

	"github.com/gin-gonic/gin"
)

// GinLogger 与zap格式一致的gin日志中间件
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		latency := end.Sub(start)

		// 获取请求信息
		path := c.Request.URL.Path
		method := c.Request.Method
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		// 使用与zap相同的格式记录日志
		if statusCode >= 400 && statusCode < 500 {
			// 警告级别
			SugaredLogger.Warnw("HTTP Request",
				"status", statusCode,
				"method", method,
				"path", path,
				"ip", clientIP,
				"latency", latency,
			)
		} else if statusCode >= 500 {
			// 错误级别
			SugaredLogger.Errorw("HTTP Request",
				"status", statusCode,
				"method", method,
				"path", path,
				"ip", clientIP,
				"latency", latency,
			)
		} else {
			// 信息级别
			SugaredLogger.Infow("HTTP Request",
				"status", statusCode,
				"method", method,
				"path", path,
				"ip", clientIP,
				"latency", latency,
			)
		}
	}
}
