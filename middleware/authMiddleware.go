package middleware // 更新包名

import (
	"crypto/des"
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
		//svrsubid := userInfo["subid"]
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
			log.Errorf("Failed to base64 decode token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// 使用hexDecodedToken对base64DecodedToken进行DES解码
		// 确保hexDecodedToken长度为8字节（DES密钥长度）
		if len(hexDecodedToken) != 8 {
			log.Errorf("Invalid DES key length: %d, must be 8 bytes", len(hexDecodedToken))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// 创建DES解密器
		desBlock, err := des.NewCipher(hexDecodedToken)
		if err != nil {
			log.Errorf("Failed to create DES cipher: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// 确保密文长度是8的倍数
		if len(base64DecodedToken)%8 != 0 {
			log.Errorf("Invalid ciphertext length: %d, must be multiple of 8", len(base64DecodedToken))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// 使用ECB模式解密（如果前8个字符正确但后面乱码，可能加密方使用了ECB模式）
		// 注意：ECB模式不安全，但有些旧系统可能仍在使用
		plaintext := make([]byte, len(base64DecodedToken))
		for i := 0; i < len(base64DecodedToken); i += des.BlockSize {
			desBlock.Decrypt(plaintext[i:i+des.BlockSize], base64DecodedToken[i:i+des.BlockSize])
		}

		// 去除ISO7816-4填充
		// ISO7816-4填充规则: 第一个字节是0x80，后面跟着0个或多个0x00字节
		// 查找0x80的位置
		paddingIndex := -1
		for i := len(plaintext) - 1; i >= 0; i-- {
			if plaintext[i] == 0x80 {
				paddingIndex = i
				break
			} else if plaintext[i] != 0x00 {
				// 如果遇到非0x00且非0x80的字节，则没有使用ISO7816-4填充
				paddingIndex = len(plaintext)
				break
			}
		}

		// 如果没有找到0x80，则假设没有填充
		if paddingIndex == -1 {
			paddingIndex = len(plaintext)
		}

		plaintext = plaintext[:paddingIndex]

		log.Info("DES decrypted data length: ", len(plaintext))
		log.Info("DES decrypted data: ", string(plaintext))

		// 可以将解密后的信息存储在上下文中
		// c.Set("subid", svrsubid)
		// c.Set("decryptedData", string(plaintext))

		// 验签通过，继续处理请求
		c.Next()
	}
}
