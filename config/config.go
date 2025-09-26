package config

import (
	"os"
	"strconv"

	"github.com/spf13/viper"
)

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault 获取环境变量整数值
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// WechatInfo 微信配置信息
type WechatInfo struct {
	ID     int    `mapstructure:"id"`
	AppID  string `mapstructure:"appid"`
	Secret string `mapstructure:"secret"`
}

// AppConfig 应用配置结构体
var AppConfig struct {
	Server struct {
		Port string
	}
	MySQL struct {
		Host     string
		Port     string
		Username string
		Password string
		Database string
		Charset  string
	}
	// gameWeb库配置 - 管理员数据
	MySQLGameWeb struct {
		Host     string
		Port     string
		Username string
		Password string
		Database string
		Charset  string
	}
	// root库配置 - 日志数据
	MySQLGameLog struct {
		Host     string
		Port     string
		Username string
		Password string
		Database string
		Charset  string
	}
	Redis struct {
		Host     string
		Port     string
		Password string
		Database int
	}
	Log struct {
		Level      string
		Path       string
		Format     string
		DateFormat string // 添加日期格式配置
	}
	// 客户端JWT配置
	JWT struct {
		SecretKey  string
		ExpireTime int64 // 过期时间，单位：秒
	}
	// 管理后台JWT配置
	Admin struct {
		JWTSecretKey     string
		TokenExpireHours int
		SessionTimeout   int
		MaxLoginAttempts int
		LockoutDuration  int
	}
	// 添加GameServer配置
	GameServer struct {
		Host string
		Port string
	}
	// 添加WechatInfo配置
	WechatInfos []WechatInfo `mapstructure:"wechatInfo"`
}

// InitConfig 初始化配置
func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// 设置默认值（优先使用环境变量）
	viper.SetDefault("Server.Port", getEnvOrDefault("SERVER_PORT", "8080"))
	viper.SetDefault("MySQL.Host", getEnvOrDefault("MYSQL_HOST", "localhost"))
	viper.SetDefault("MySQL.Port", getEnvOrDefault("MYSQL_PORT", "3306"))
	viper.SetDefault("MySQL.Username", getEnvOrDefault("MYSQL_USER", "root"))
	viper.SetDefault("MySQL.Password", getEnvOrDefault("MYSQL_PASSWORD", "password"))
	viper.SetDefault("MySQL.Database", getEnvOrDefault("MYSQL_DATABASE", "game_db"))
	viper.SetDefault("MySQL.Charset", "utf8mb4")
	// 设置gameWeb数据库默认值
	viper.SetDefault("MySQLGameWeb.Host", getEnvOrDefault("MYSQL_GAMEWEB_HOST", "localhost"))
	viper.SetDefault("MySQLGameWeb.Port", getEnvOrDefault("MYSQL_GAMEWEB_PORT", "3306"))
	viper.SetDefault("MySQLGameWeb.Username", getEnvOrDefault("MYSQL_GAMEWEB_USER", "root"))
	viper.SetDefault("MySQLGameWeb.Password", getEnvOrDefault("MYSQL_GAMEWEB_PASSWORD", "password"))
	viper.SetDefault("MySQLGameWeb.Database", getEnvOrDefault("MYSQL_GAMEWEB_DATABASE", "gameWeb"))
	viper.SetDefault("MySQLGameWeb.Charset", "utf8mb4")
	// 设置root数据库默认值（优先使用环境变量）
	viper.SetDefault("MySQLGameLog.Host", getEnvOrDefault("MYSQL_GAMELOG_HOST", "localhost"))
	viper.SetDefault("MySQLGameLog.Port", getEnvOrDefault("MYSQL_GAMELOG_PORT", "3306"))
	viper.SetDefault("MySQLGameLog.Username", getEnvOrDefault("MYSQL_GAMELOG_USER", "root"))
	viper.SetDefault("MySQLGameLog.Password", getEnvOrDefault("MYSQL_GAMELOG_PASSWORD", "password"))
	viper.SetDefault("MySQLGameLog.Database", getEnvOrDefault("MYSQL_GAMELOG_DATABASE", "root"))
	viper.SetDefault("MySQLGameLog.Charset", "utf8mb4")
	// 其他配置默认值...
	viper.SetDefault("Redis.Host", getEnvOrDefault("REDIS_HOST", "localhost"))
	viper.SetDefault("Redis.Port", getEnvOrDefault("REDIS_PORT", "6379"))
	viper.SetDefault("Redis.Password", getEnvOrDefault("REDIS_PASSWORD", ""))
	viper.SetDefault("Redis.Database", getEnvIntOrDefault("REDIS_DATABASE", 0))
	viper.SetDefault("Log.Level", "info")
	viper.SetDefault("Log.Path", "logs/game.log")
	viper.SetDefault("Log.Format", "console")
	viper.SetDefault("Log.DateFormat", "2006-01-02") // 添加默认日期格式
	// 添加客户端JWT默认值
	viper.SetDefault("JWT.SecretKey", getEnvOrDefault("JWT_SECRET", "GameWebJWTSecretKey1234567890ABCDEF"))
	viper.SetDefault("JWT.ExpireTime", getEnvIntOrDefault("JWT_EXPIRE_TIME", 3600)) // 默认1小时过期
	// 添加管理后台JWT默认值
	viper.SetDefault("Admin.JWTSecretKey", getEnvOrDefault("ADMIN_JWT_SECRET", "GameWebAdminJWTSecretKey987654321FEDCBA"))
	viper.SetDefault("Admin.TokenExpireHours", getEnvIntOrDefault("ADMIN_TOKEN_EXPIRE_HOURS", 8)) // 8小时过期
	viper.SetDefault("Admin.SessionTimeout", getEnvIntOrDefault("ADMIN_SESSION_TIMEOUT", 24))     // Redis会话24小时过期
	viper.SetDefault("Admin.MaxLoginAttempts", getEnvIntOrDefault("ADMIN_MAX_LOGIN_ATTEMPTS", 5)) // 最大登录尝试次数
	viper.SetDefault("Admin.LockoutDuration", getEnvIntOrDefault("ADMIN_LOCKOUT_DURATION", 30))   // 锁定时间（分钟）
	// 添加GameServer默认值
	viper.SetDefault("GameServer.Host", getEnvOrDefault("GAMESERVER_HOST", "localhost"))
	viper.SetDefault("GameServer.Port", getEnvOrDefault("GAMESERVER_PORT", "9000"))

	// 添加WechatInfo默认值
	viper.SetDefault("wechatInfo", []map[string]interface{}{
		{"id": 1, "appid": "", "secret": ""},
	})

	if err := viper.ReadInConfig(); err != nil {
		panic("Failed to read config file: " + err.Error())
	}

	// 绑定配置到结构体
	if err := viper.Unmarshal(&AppConfig); err != nil {
		panic("Failed to unmarshal config: " + err.Error())
	}
}
