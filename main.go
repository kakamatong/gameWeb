package main

// 在导入部分添加CORS相关包
import (
	"gameWeb/config"
	"gameWeb/db"
	"gameWeb/routes"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func init() {
	// 初始化配置
	config.InitConfig()

	// 初始化日志系统
	config.InitLogger()

	// 初始化数据库连接
	db.InitMySQL()
	db.InitRedis()
}

func main() {
	logrus.Info("Starting game web API server")

	// 设置gin为发布模式，关闭debug
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	router := gin.Default()

	// 添加CORS中间件
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},                                       // 允许所有来源，生产环境应限制具体域名
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // 允许的HTTP方法
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"}, // 允许的请求头
		ExposeHeaders:    []string{"Content-Length"},                          // 暴露的响应头
		AllowCredentials: true,                                                // 允许携带Cookie
		MaxAge:           12 * time.Hour,                                      // 预检请求的有效期
	}))

	// 注册路由
	routes.RegisterRoutes(router)

	// 启动服务器
	serverPort := config.AppConfig.Server.Port
	if err := router.Run(":" + serverPort); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}
