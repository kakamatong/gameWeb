package middleware // 更新包名

import (
	"encoding/base64"
	"encoding/hex"
	"gameWeb/db"
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

		// 验证token格式
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
		log.Info("token ", token, " userid ", userid)
		key := "user:" + userid

		// 使用HGetAll从Redis中获取用户信息
		userInfo, err := db.HGetAllRedis(key)
		if err != nil {
			log.Errorf("Failed to get user info from Redis: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}
		log.Info("userInfo ", userInfo)

		// 检查用户信息是否存在
		if len(userInfo) == 0 {
			log.Errorf("User info not found in Redis for user: %s", userid)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// 获取subid和token
		svrsubid := userInfo["subid"]
		svrtoken := userInfo["token"]

		// 对svrtoken进行hex解码
		hexDecodedToken, err := hex.DecodeString(svrtoken)
		if err != nil {
			log.Errorf("Failed to hex decode svrtoken: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}
		log.Info("Hex decoded svrtoken: ", string(hexDecodedToken))

		// 对token进行base64解码
		base64DecodedToken, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			// 尝试使用URL安全的base64解码
			base64DecodedToken, err = base64.URLEncoding.DecodeString(token)
			if err != nil {
				log.Errorf("Failed to base64 decode token: %v", err)
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "Invalid or expired token",
				})
				c.Abort()
				return
			}
		}
		log.Info("Base64 decoded token: ", string(base64DecodedToken))

		// 验签通过，继续处理请求
		c.Next()
	}
}
