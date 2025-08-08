package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

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
	Redis struct {
		Host     string
		Port     string
		Password string
		Database int
	}
	Log struct {
		Level  string
		Path   string
		Format string
	}
}

// InitConfig 初始化配置
func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// 设置默认值
	viper.SetDefault("Server.Port", "8080")
	viper.SetDefault("MySQL.Host", "localhost")
	viper.SetDefault("MySQL.Port", "3306")
	viper.SetDefault("MySQL.Username", "root")
	viper.SetDefault("MySQL.Password", "password")
	viper.SetDefault("MySQL.Database", "game_db")
	viper.SetDefault("MySQL.Charset", "utf8mb4")
	viper.SetDefault("Redis.Host", "localhost")
	viper.SetDefault("Redis.Port", "6379")
	viper.SetDefault("Redis.Password", "")
	viper.SetDefault("Redis.Database", 0)
	viper.SetDefault("Log.Level", "info")
	viper.SetDefault("Log.Path", "logs/game.log")
	viper.SetDefault("Log.Format", "text")

	if err := viper.ReadInConfig(); err != nil {
		logrus.Warnf("Failed to read config file: %v, using default values", err)
	}

	// 绑定配置到结构体
	if err := viper.Unmarshal(&AppConfig); err != nil {
		logrus.Fatalf("Failed to unmarshal config: %v", err)
	}
}
