package log

import (
    "os"
    "path/filepath"

    "github.com/sirupsen/logrus"
    "gameWeb/config"
)

// Logger 全局日志对象
var Logger = logrus.New()

// InitLogger 初始化日志系统
func InitLogger() {
    // 设置日志级别
    level, err := logrus.ParseLevel(config.AppConfig.Log.Level)
    if err != nil {
        level = logrus.InfoLevel
    }
    Logger.SetLevel(level)

    // 设置日志格式
    if config.AppConfig.Log.Format == "json" {
        Logger.SetFormatter(&logrus.JSONFormatter{
            TimestampFormat: "2006-01-02 15:04:05",
        })
    } else {
        Logger.SetFormatter(&logrus.TextFormatter{
            TimestampFormat: "2006-01-02 15:04:05",
            FullTimestamp:   true,
        })
    }

    // 设置日志输出路径
    if config.AppConfig.Log.Path != "" {
        // 确保目录存在
        dir := filepath.Dir(config.AppConfig.Log.Path)
        if err := os.MkdirAll(dir, 0755); err != nil {
            Logger.Fatalf("Failed to create log directory: %v", err)
        }

        // 创建日志文件
        file, err := os.OpenFile(
            config.AppConfig.Log.Path,
            os.O_APPEND|os.O_CREATE|os.O_WRONLY,
            0644,
        )
        if err != nil {
            Logger.Fatalf("Failed to open log file: %v", err)
        }

        Logger.SetOutput(file)
    }
}

// 日志级别常量
const (
    LogLevelDebug = "debug"
    LogLevelInfo  = "info"
    LogLevelWarn  = "warn"
    LogLevelError = "error"
    LogLevelFatal = "fatal"
)