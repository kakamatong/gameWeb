package db

import (
	"database/sql"
	"fmt"
	"gameWeb/config"

	"github.com/sirupsen/logrus"
)

var MySQLDB *sql.DB

// InitMySQL 初始化MySQL连接
func InitMySQL() {
	cfg := config.AppConfig.MySQL

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
	)

	// 打开数据库连接
	var err error
	MySQLDB, err = sql.Open("mysql", dsn)
	if err != nil {
		logrus.Fatalf("Failed to open MySQL connection: %v", err)
	}

	// 测试连接
	if err := MySQLDB.Ping(); err != nil {
		logrus.Fatalf("Failed to ping MySQL: %v", err)
	}

	logrus.Info("Successfully connected to MySQL")

	// 设置连接池参数
	MySQLDB.SetMaxOpenConns(100)
	MySQLDB.SetMaxIdleConns(20)
	MySQLDB.SetConnMaxLifetime(3600)
}