package log

import (
    "time"

    "github.com/gin-gonic/gin"
)

// GinLogger 自定义Gin日志中间件
func GinLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        method := c.Request.Method

        // 处理请求
        c.Next()

        // 计算请求时间
        latency := time.Since(start)
        statusCode := c.Writer.Status()

        // 构建日志字段
        fields := map[string]interface{}{
            "status":     statusCode,
            "method":     method,
            "path":       path,
            "latency":    latency,
            "client_ip":  c.ClientIP(),
            "user_agent": c.Request.UserAgent(),
        }

        // 根据状态码记录不同级别的日志
        if statusCode >= 500 {
            Logger.WithFields(fields).Error("Server error")
        } else if statusCode >= 400 {
            Logger.WithFields(fields).Warn("Client error")
        } else {
            Logger.WithFields(fields).Info("Request processed")
        }
    }
}

// GinRecovery 自定义Gin恢复中间件
func GinRecovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // 记录错误日志
                Logger.WithField("error", err).Error("Panic recovered")

                // 返回500错误
                c.JSON(500, gin.H{
                    "code": 500,
                    "msg":  "Internal server error",
                })

                // 终止请求
                c.Abort()
            }
        }()

        c.Next()
    }
}
