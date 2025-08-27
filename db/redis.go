package db

import (
	"context"
	"fmt"
	"gameWeb/config"
	"gameWeb/log"
	"time"

	"github.com/go-redis/redis/v8"
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
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Info("Successfully connected to Redis")
}

// SetRedis 设置Redis键值对
func SetRedis(key string, value interface{}, expiration time.Duration) error {
	return RedisClient.Set(ctx, key, value, expiration).Err()
}

// GetRedis 获取Redis键值对
func GetRedis(key string) (string, error) {
	return RedisClient.Get(ctx, key).Result()
}

// HGetAllRedis 获取哈希表中所有字段和值
func HGetAllRedis(key string) (map[string]string, error) {
	return RedisClient.HGetAll(ctx, key).Result()
}

// SetRedisWithExpire 设置Redis键值对并设置过期时间
func SetRedisWithExpire(key string, value interface{}, expiration time.Duration) error {
	return RedisClient.Set(ctx, key, value, expiration).Err()
}

// DelRedis 删除Redis键
func DelRedis(key string) error {
	return RedisClient.Del(ctx, key).Err()
}

// ExistsRedis 检查Redis键是否存在
func ExistsRedis(key string) (bool, error) {
	result, err := RedisClient.Exists(ctx, key).Result()
	return result > 0, err
}

// ExpireRedis 设置Redis键的过期时间
func ExpireRedis(key string, expiration time.Duration) error {
	return RedisClient.Expire(ctx, key, expiration).Err()
}
