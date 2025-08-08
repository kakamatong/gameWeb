package db

import (
	"context"
	"fmt"
	"gameWeb/config"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

var RedisClient *redis.Client
var ctx = context.Background()

// InitRedis 初始化Redis连接
func InitRedis() {
	cfg := config.AppConfig.Redis

	// 创建Redis客户端
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.Database,
	})

	// 测试连接
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		logrus.Fatalf("Failed to connect to Redis: %v", err)
	}

	logrus.Info("Successfully connected to Redis")
}

// SetRedis 设置Redis键值对
func SetRedis(key string, value interface{}, expiration time.Duration) error {
	return RedisClient.Set(ctx, key, value, expiration).Err()
}

// GetRedis 获取Redis键值对
func GetRedis(key string) (string, error) {
	return RedisClient.Get(ctx, key).Result()
}
