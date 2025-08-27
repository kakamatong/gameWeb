package middleware

import (
	"crypto/des"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gameWeb/config"
	"gameWeb/db"
	"gameWeb/log"
	"gameWeb/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// TokenInfo 定义存储JSON数据的结构体
type TokenInfo struct {
	Userid int64 `json:"userid"`
	Subid  int64 `json:"subid"`
	Time   int64 `json:"time"`
}

// AuthMiddleware 生成验签中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从请求头获取认证信息
		auth := c.GetHeader("Authorization")
		userid := c.GetHeader("X-User-ID")

		// 验证必要头信息是否存在
		if auth == "" || userid == "" {
			log.Errorf("缺少Authorization或X-User-ID")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少Authorization或X-User-ID",
			})
			c.Abort()
			return
		}

		// 2. 验证token格式
		if !strings.HasPrefix(auth, "Bearer ") {
			log.Errorf("Authorization格式无效")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的Authorization格式",
			})
			c.Abort()
			return
		}

		// 提取token
		token := auth[7:]
		log.Info("token ", token, " userid ", userid)
		key := "user:" + userid

		// 3. 从Redis中获取用户信息
		userInfo, err := db.HGetAllRedis(key)
		if err != nil {
			log.Errorf("从Redis获取用户信息失败: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效或过期的token",
			})
			c.Abort()
			return
		}
		log.Info("userInfo ", userInfo)

		// 检查用户信息是否存在
		if len(userInfo) == 0 {
			log.Errorf("Redis中未找到用户信息: %s", userid)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效或过期的token",
			})
			c.Abort()
			return
		}

		// 获取subid和token
		svrsubid := userInfo["subid"]
		svrtoken := userInfo["token"]

		// 4. 对svrtoken进行hex解码
		hexDecodedToken, err := hex.DecodeString(svrtoken)
		if err != nil {
			log.Errorf("svrtoken hex解码失败: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效或过期的token",
			})
			c.Abort()
			return
		}
		log.Info("Hex解码后的svrtoken: ", string(hexDecodedToken))

		// 5. 对token进行base64解码
		base64DecodedToken, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			log.Errorf("token base64解码失败: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效或过期的token",
			})
			c.Abort()
			return
		}

		// 6. 使用DES算法解密
		// 确保DES密钥长度为8字节
		if len(hexDecodedToken) != 8 {
			log.Errorf("DES密钥长度无效: %d, 必须为8字节", len(hexDecodedToken))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效或过期的token",
			})
			c.Abort()
			return
		}

		// 创建DES解密器
		desBlock, err := des.NewCipher(hexDecodedToken)
		if err != nil {
			log.Errorf("创建DES解密器失败: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效或过期的token",
			})
			c.Abort()
			return
		}

		// 确保密文长度是8的倍数
		if len(base64DecodedToken)%8 != 0 {
			log.Errorf("密文长度无效: %d, 必须是8的倍数", len(base64DecodedToken))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效或过期的token",
			})
			c.Abort()
			return
		}

		// 使用ECB模式解密
		plaintext := make([]byte, len(base64DecodedToken))
		for i := 0; i < len(base64DecodedToken); i += des.BlockSize {
			desBlock.Decrypt(plaintext[i:i+des.BlockSize], base64DecodedToken[i:i+des.BlockSize])
		}

		// 7. 去除ISO7816-4填充
		// ISO7816-4填充规则: 第一个字节是0x80，后面跟着0个或多个0x00字节
		paddingIndex := -1
		for i := len(plaintext) - 1; i >= 0; i-- {
			if plaintext[i] == 0x80 {
				paddingIndex = i
				break
			} else if plaintext[i] != 0x00 {
				// 遇到非0x00且非0x80的字节，没有使用ISO7816-4填充
				paddingIndex = len(plaintext)
				break
			}
		}

		// 如果没有找到0x80，则假设没有填充
		if paddingIndex == -1 {
			paddingIndex = len(plaintext)
		}

		plaintext = plaintext[:paddingIndex]

		log.Info("DES解密数据长度: ", len(plaintext))
		log.Info("DES解密数据: ", string(plaintext))

		// 8. JSON解析plaintext数据
		var tokenInfo TokenInfo
		err = json.Unmarshal(plaintext, &tokenInfo)
		if err != nil {
			log.Errorf("解析token信息失败: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的token格式",
			})
			c.Abort()
			return
		}

		// 9. 验证userid和subid
		// 转换请求头中的userid为int64类型
		reqUserid, err := strconv.ParseInt(userid, 10, 64)
		if err != nil {
			log.Errorf("userid格式无效: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的userid格式",
			})
			c.Abort()
			return
		}

		// 转换Redis中的subid为int64类型
		svrSubid, err := strconv.ParseInt(svrsubid, 10, 64)
		if err != nil {
			log.Errorf("Redis中的subid格式无效: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的token数据",
			})
			c.Abort()
			return
		}

		// 比较解析出的userid和subid
		if tokenInfo.Userid != reqUserid || tokenInfo.Subid != svrSubid {
			log.Errorf("token验证失败: 期望userid=%d, subid=%d; 实际userid=%d, subid=%d",
				reqUserid, svrSubid, tokenInfo.Userid, tokenInfo.Subid)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效或过期的token",
			})
			c.Abort()
			return
		}

		log.Info("用户token验证成功: ", userid)

		// 10. 将验证后的信息存储在上下文中
		c.Set("subid", tokenInfo.Subid)
		c.Set("userid", tokenInfo.Userid)
		c.Set("tokenTime", tokenInfo.Time)

		// 验签通过，继续处理请求
		c.Next()
	}
}

// JWTClaims 定义JWT声明结构体
type JWTClaims struct {
	Userid    int64  `json:"userid"`
	Channelid string `json:"channelid"`
	jwt.RegisteredClaims
}

// AuthMiddlewareByJWT 基于JWT的认证中间件（简化版，无需Redis验证）
func AuthMiddlewareByJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从请求头获取认证信息
		auth := c.GetHeader("Authorization")

		// 验证必要头信息是否存在
		if auth == "" {
			log.Errorf("缺少Authorization")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少Authorization",
			})
			c.Abort()
			return
		}

		// 2. 验证token格式
		if !strings.HasPrefix(auth, "Bearer ") {
			log.Errorf("Authorization格式无效")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的Authorization格式",
			})
			c.Abort()
			return
		}

		// 提取token
		tokenString := auth[7:]
		log.Info("JWT token: ", tokenString)
		// 3. 解析和验证JWT token
		secretKey := []byte(config.AppConfig.JWT.SecretKey)
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secretKey, nil
		})

		if err != nil {
			log.Errorf("JWT解析失败: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效或过期的token",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// 验证token是否有效
		if !token.Valid {
			log.Errorf("token无效")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的token",
			})
			c.Abort()
			return
		}

		// 提取claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			log.Errorf("无法提取JWT claims")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的token格式",
			})
			c.Abort()
			return
		}

		// 4. 将验证后的信息存储在上下文中
		c.Set("userid", claims.Userid)
		c.Set("channelid", claims.Channelid)
		c.Set("tokenTime", claims.IssuedAt.Unix())

		log.Info("用户JWT验证成功: userid=", claims.Userid, ", channelid=", claims.Channelid)

		// 验签通过，继续处理请求
		c.Next()
	}
}

// AdminJWTMiddleware 管理员专用JWT认证中间件
func AdminJWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 提取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少Authorization头",
			})
			c.Abort()
			return
		}

		// 2. 验证Bearer格式
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的Authorization格式",
			})
			c.Abort()
			return
		}

		// 3. 提取并验证管理员JWT Token
		tokenString := authHeader[7:]
		claims, err := ValidateAdminJWT(tokenString)
		if err != nil {
			log.Errorf("管理员JWT验证失败: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效或过期的token",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// 4. 检查管理员会话状态
		sessionKey := fmt.Sprintf("admin_session:%d", claims.AdminID)
		sessionData, err := db.GetRedis(sessionKey)
		if err != nil || sessionData == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "会话已过期，请重新登录",
			})
			c.Abort()
			return
		}

		// 5. 验证IP地址（可选安全检查）
		clientIP := c.ClientIP()
		if claims.LoginIP != "" && claims.LoginIP != clientIP {
			log.Warnf("管理员IP地址变更: %s -> %s, 用户: %s", claims.LoginIP, clientIP, claims.Username)
			// 可以选择是否严格验证IP，这里只记录警告
		}

		// 6. 存储管理员信息到上下文
		c.Set("adminId", claims.AdminID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("realName", claims.RealName)
		c.Set("isSuperAdmin", claims.IsSuperAdmin)
		c.Set("departmentId", claims.DepartmentID)
		c.Set("loginIp", claims.LoginIP)
		
		log.Infof("管理员认证成功: %s (ID: %d)", claims.Username, claims.AdminID)
		c.Next()
	}
}

// ValidateAdminJWT 验证管理员JWT Token
func ValidateAdminJWT(tokenString string) (*models.AdminJWTClaims, error) {
	// 使用专门的管理员JWT密钥
	secretKey := []byte(config.AppConfig.Admin.JWTSecretKey)
	
	token, err := jwt.ParseWithClaims(tokenString, &models.AdminJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("JWT解析失败: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token无效")
	}

	claims, ok := token.Claims.(*models.AdminJWTClaims)
	if !ok {
		return nil, fmt.Errorf("无法提取JWT claims")
	}

	return claims, nil
}

// GenerateAdminJWT 生成管理员JWT Token
func GenerateAdminJWT(admin *models.AdminAccount) (string, error) {
	now := time.Now()
	expirationTime := now.Add(time.Duration(config.AppConfig.Admin.TokenExpireHours) * time.Hour)

	claims := &models.AdminJWTClaims{
		AdminID:      admin.ID,
		Username:     admin.Username,
		Email:        admin.Email,
		RealName:     admin.RealName,
		IsSuperAdmin: admin.IsSuperAdmin == 1,
		DepartmentID: admin.DepartmentID,
		LoginIP:      admin.LastLoginIP,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "gameWeb-admin",
			Subject:   fmt.Sprintf("admin:%d", admin.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secretKey := []byte(config.AppConfig.Admin.JWTSecretKey)
	
	return token.SignedString(secretKey)
}

// RequireSuperAdmin 需要超级管理员权限中间件
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		isSuperAdmin, exists := c.Get("isSuperAdmin")
		if !exists || !isSuperAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "需要超级管理员权限",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
