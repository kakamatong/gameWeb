package log

import (
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger 全局日志对象
var Logger = logrus.New()

// LogConfig 日志配置结构体
type LogConfig struct {
	Level  string
	Path   string
	Format string
}

// InitLogger 初始化日志系统
func InitLogger(config LogConfig) {
	// 设置日志级别
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)

	// 设置日志格式
	if config.Format == "json" {
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		Logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
		})
	}

	// 设置日志输出
	if config.Path != "" {
		// 确保目录存在
		dir := filepath.Dir(config.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			Logger.Fatalf("Failed to create log directory: %v", err)
		}

		// 使用按小时分割的日志写入器
		Logger.SetOutput(&HourlyRotatingFileWriter{
			BasePath: config.Path,
		})
	}
}

// HourlyRotatingFileWriter 按小时分割日志文件
type HourlyRotatingFileWriter struct {
	BasePath    string
	currentFile *os.File
	currentHour time.Time
}

// Write 实现 io.Writer 接口
func (w *HourlyRotatingFileWriter) Write(p []byte) (n int, err error) {
	now := time.Now()
	hourKey := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())

	// 如果当前小时与记录的小时不同，或者文件未打开，则创建新文件
	if w.currentFile == nil || !hourKey.Equal(w.currentHour) {
		if w.currentFile != nil {
			w.currentFile.Close()
		}

		// 生成带小时的文件名
		fileName := w.BasePath + "." + hourKey.Format("2006010215")
		file, err := os.OpenFile(
			fileName,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0644,
		)
		if err != nil {
			return 0, err
		}

		w.currentFile = file
		w.currentHour = hourKey
	}

	return w.currentFile.Write(p)
}

// Close 关闭日志文件
func (w *HourlyRotatingFileWriter) Close() error {
	if w.currentFile != nil {
		return w.currentFile.Close()
	}
	return nil
}

// 日志级别常量
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)
