// redis.go 提供 Redis 缓存连接初始化功能。
//
// Redis 作为可选组件：连接失败时记录警告日志但不中断主流程，
// 系统可在无缓存模式下正常运行。
package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-redis/redis/v8"

	"casbin-demo/internal/config"
)

// RedisClient 全局 Redis 客户端实例。
// 在 InitRedis() 成功后初始化；若 Redis 不可用则为 nil。
var RedisClient *redis.Client

// Ctx 全局 Redis 操作上下文。
// 注意：生产环境应使用请求级 context 而非全局 context。
var Ctx = context.Background()

// InitRedis 初始化 Redis 连接。
//
// Redis 为可选组件：连接失败时返回 error 但不影响主流程，
// 上层 main() 会记录 Warn 日志并在无缓存模式下继续运行。
//
// 配置参数：
//   - Addr: host:port
//   - Password: 密码（无密码留空）
//   - DB: 数据库编号（默认 0）
//   - PoolSize: 连接池大小
func InitRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	if err := client.Ping(Ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	slog.Info("redis connected",
		"host", cfg.Redis.Host,
		"port", cfg.Redis.Port,
		"db", cfg.Redis.DB,
	)
	RedisClient = client
	return client, nil
}
