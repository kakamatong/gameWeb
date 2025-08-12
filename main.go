package main

import (
	"gameWeb/config"
	"gameWeb/db"
	"gameWeb/log"
	"gameWeb/routes"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func init() {
	config.InitConfig()

	// 初始化日志系统，传递配置
	log.InitZapLog(log.LogConfig{
		Level:  config.AppConfig.Log.Level,
		Path:   config.AppConfig.Log.Path,
		Format: config.AppConfig.Log.Format,
	})

	// 初始化数据库连接
	db.InitMySQL()
	db.InitMySQLGameWeb() // 初始化第二个数据库连接
	db.InitRedis()
}

func main() {

	// 设置gin为发布模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎，禁用默认日志
	router := gin.New()

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
		log.Fatalf("Failed to start server: %v", err)
	}
}
