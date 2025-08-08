package config

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GinLogger 自定义Gin日志中间件
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// 请求信息
		reqMethod := c.Request.Method
		reqURI := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		// 根据日志格式输出
		if AppConfig.Log.Format == "json" {
			logrus.WithFields(logrus.Fields{
				"method":  reqMethod,
				"uri":     reqURI,
				"status":  statusCode,
				"ip":      clientIP,
				"latency": latency,
				// 移除手动添加的timestamp字段
			}).Info("HTTP Request")
		} else {
			logrus.Infof("%s %s %d %s",
				clientIP,
				reqMethod,
				statusCode,
				reqURI,
			)
		}
	}
}

// GinRecovery 自定义恢复中间件
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logrus.Errorf("Panic recovered: %v", err)
				c.JSON(500, gin.H{
					"code":    500,
					"message": "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
