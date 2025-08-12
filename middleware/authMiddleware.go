package middleware // 更新包名

import (
	"gameWeb/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 生成验签中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从header中获取签名信息
		signature := c.GetHeader("X-Signature")

		// 检查签名和时间戳是否存在
		if signature == "" {
			log.Errorf("Missing signature")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Missing signature",
			})
			c.Abort()
			return
		}

		// 验签通过，继续处理请求
		c.Next()
	}
}
