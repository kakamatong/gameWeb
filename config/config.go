package config

import (
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
	// 添加第二个数据库配置
	MySQLGameWeb struct {
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
	// 设置第二个数据库默认值
	viper.SetDefault("MySQLGameWeb.Host", "localhost")
	viper.SetDefault("MySQLGameWeb.Port", "3306")
	viper.SetDefault("MySQLGameWeb.Username", "root")
	viper.SetDefault("MySQLGameWeb.Password", "password")
	viper.SetDefault("MySQLGameWeb.Database", "gameWeb")
	viper.SetDefault("MySQLGameWeb.Charset", "utf8mb4")
	// 其他配置默认值...
	viper.SetDefault("Redis.Host", "localhost")
	viper.SetDefault("Redis.Port", "6379")
	viper.SetDefault("Redis.Password", "")
	viper.SetDefault("Redis.Database", 0)
	viper.SetDefault("Log.Level", "info")
	viper.SetDefault("Log.Path", "logs/game.log")
	viper.SetDefault("Log.Format", "console")
	viper.SetDefault("Log.DateFormat", "2006-01-02") // 添加默认日期格式

	if err := viper.ReadInConfig(); err != nil {
		panic("Failed to read config file: " + err.Error())
	}

	// 绑定配置到结构体
	if err := viper.Unmarshal(&AppConfig); err != nil {
		panic("Failed to unmarshal config: " + err.Error())
	}
}
