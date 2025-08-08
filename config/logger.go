package config

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// InitLogger 初始化日志系统
func InitLogger() {
	// 设置日志级别
	level, err := logrus.ParseLevel(AppConfig.Log.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// 设置日志格式
	if AppConfig.Log.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
		})
	}

	// 设置日志输出路径
	if AppConfig.Log.Path != "" {
		// 确保目录存在
		dir := filepath.Dir(AppConfig.Log.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			logrus.Fatalf("Failed to create log directory: %v", err)
		}

		// 创建日志文件
		file, err := os.OpenFile(
			AppConfig.Log.Path,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0644,
		)
		if err != nil {
			logrus.Fatalf("Failed to open log file: %v", err)
		}

		logrus.SetOutput(file)
	}
}
