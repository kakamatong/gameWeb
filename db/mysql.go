package db

import (
	"database/sql"
	"fmt"
	"gameWeb/config"
	"gameWeb/log"

	// Add the MySQL driver import
	_ "github.com/go-sql-driver/mysql"
)

var MySQLDB *sql.DB
var MySQLDBGameWeb *sql.DB // 第二个数据库连接

// InitMySQL 初始化第一个MySQL连接
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
		log.Fatalf("Failed to open MySQL connection: %v", err)
	}

	// 测试连接
	if err := MySQLDB.Ping(); err != nil {
		log.Fatalf("Failed to ping MySQL: %v", err)
	}

	log.Info("Successfully connected to MySQL")

	// 设置连接池参数
	MySQLDB.SetMaxOpenConns(100)
	MySQLDB.SetMaxIdleConns(20)
	MySQLDB.SetConnMaxLifetime(3600)
}

// InitMySQLGameWeb 初始化第二个MySQL连接 (gameWeb数据库)
func InitMySQLGameWeb() {
	cfg := config.AppConfig.MySQLGameWeb

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
	MySQLDBGameWeb, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open MySQLGameWeb connection: %v", err)
	}

	// 测试连接
	if err := MySQLDBGameWeb.Ping(); err != nil {
		log.Fatalf("Failed to ping MySQLGameWeb: %v", err)
	}

	log.Info("Successfully connected to MySQLGameWeb")

	// 设置连接池参数
	MySQLDBGameWeb.SetMaxOpenConns(100)
	MySQLDBGameWeb.SetMaxIdleConns(20)
	MySQLDBGameWeb.SetConnMaxLifetime(3600)
}
