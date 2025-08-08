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

	// 设置gin为发布模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎，禁用默认日志
	router := gin.New()

	// 使用自定义日志中间件 - 这里的引用应该已经是正确的，因为我们引用的是包而不是具体文件
	router.Use(config.GinLogger(), config.GinRecovery())

	// 添加CORS中间件
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 注册路由
	routes.RegisterRoutes(router)

	// 启动服务器
	serverPort := config.AppConfig.Server.Port
	if err := router.Run(":" + serverPort); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}
