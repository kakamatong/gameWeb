package log

import (
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger
var SugaredLogger *zap.SugaredLogger
var LogWriter io.Writer // 导出日志写入器

// LogConfig 日志配置结构体
type LogConfig struct {
	Level  string
	Path   string
	Format string
}

// InitZapLog 初始化zap日志系统
func InitZapLog(config LogConfig) {
	// 创建日志目录
	logDir := filepath.Dir(config.Path)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.MkdirAll(logDir, 0755)
		if err != nil {
			panic("Failed to create log directory: " + err.Error())
		}
	}

	// 设置日志级别
	var level zapcore.Level
	switch config.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "dpanic":
		level = zapcore.DPanicLevel
	case "panic":
		level = zapcore.PanicLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	// 创建编码器
	var encoder zapcore.Encoder
	if config.Format == "json" {
		encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	} else {
		encoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	}

	// 创建滚动文件写入器
	writer := &lumberjack.Logger{
		Filename:   config.Path,
		MaxSize:    100,  // 单个文件最大100MB
		MaxBackups: 5,    // 保留5个备份
		MaxAge:     5,    // 保留5天
		Compress:   false,
	}

	LogWriter = writer // 保存写入器引用

	// 创建核心
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(writer),
		level,
	)

	// 创建日志器
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	SugaredLogger = Logger.Sugar()

	// 记录初始化信息
	Logger.Info("Zap log system initialized")
}

// 以下是封装的日志方法，保持与logrus类似的接口

func Debug(args ...interface{}) {
	SugaredLogger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	SugaredLogger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	SugaredLogger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	SugaredLogger.Infof(format, args...)
}

func Warn(args ...interface{}) {
	SugaredLogger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	SugaredLogger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	SugaredLogger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	SugaredLogger.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	SugaredLogger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	SugaredLogger.Fatalf(format, args...)
}
