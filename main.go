package main

import (
	"fmt"
	"gameWeb/config"
	"gameWeb/db"
	"gameWeb/log"
	"gameWeb/routes"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func init() {
	config.InitConfig()

	// 初始化日志系统，传递配置
	log.InitZapLog(log.LogConfig{
		Level:  config.AppConfig.Log.Level,
		Path:   generateLogFilePath(config.AppConfig.Log.Path, config.AppConfig.Log.DateFormat),
		Format: config.AppConfig.Log.Format,
	})

	// 初始化数据库连接
	db.InitMySQL()        // game库 - 用户游戏数据
	db.InitMySQLGameWeb() // gameWeb库 - 管理员数据
	db.InitMySQLGameLog() // gamelog库 - 日志数据
	db.InitRedis()
}

func main() {

	// 设置gin为发布模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	router := gin.New()

	// 删除这行代码
	// router.Use(gin.LoggerWithWriter(log.LogWriter))

	// 使用自定义的gin logger中间件
	router.Use(log.GinLogger())

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

// 添加生成带日期的文件名函数
func generateLogFilePath(basePath string, dateFormat string) string {
	if dateFormat == "" {
		dateFormat = "2006-01-02"
	}
	// 获取当前日期
	currentDate := time.Now().Format(dateFormat)
	// 提取文件名和扩展名
	dir := filepath.Dir(basePath)
	filename := filepath.Base(basePath)
	nameWithoutExt := filename[:len(filename)-len(filepath.Ext(filename))]
	ext := filepath.Ext(filename)
	// 生成带日期的文件名
	return filepath.Join(dir, fmt.Sprintf("%s_%s%s", nameWithoutExt, currentDate, ext))
}
