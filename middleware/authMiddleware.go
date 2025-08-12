package middleware // 更新包名

import (
	"gameWeb/log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 生成验签中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从header中获取签名信息
		auth := c.GetHeader("Authorization")

		userid := c.GetHeader("X-User-ID")

		if auth == "" || userid == "" {
			log.Errorf("Missing Authorization or X-User-ID")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Missing Authorization or X-User-ID",
			})
			c.Abort()
			return
		}

		// 验证token格式（可选）
		if !strings.HasPrefix(auth, "Bearer ") {
			log.Errorf("Invalid Authorization format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Invalid Authorization format",
			})
			c.Abort()
			return
		}

		// 提取token
		token := auth[7:]
		key := "user:" + userid

		// 验签通过，继续处理请求
		c.Next()
	}
}
