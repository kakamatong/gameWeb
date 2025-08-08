package main

import (
	"gameWeb/config"
	"gameWeb/db"
	"gameWeb/routes"
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

	// 创建Gin引擎
	router := gin.Default()

	// 注册路由
	routes.RegisterRoutes(router)

	// 启动服务器
	serverPort := config.AppConfig.Server.Port
	if err := router.Run(":" + serverPort); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}