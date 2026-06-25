// Package config 负责应用配置的加载与默认值管理。
//
// 配置来源优先级（从高到低）：
//  1. 环境变量（DATABASE_HOST / REDIS_HOST / SERVER_PORT 等）
//  2. YAML 配置文件（config.yaml 或 config-docker.yaml）
//  3. 代码内置默认值（Default() 函数提供）
//
// 使用 viper 作为配置库，支持 . 分隔的嵌套键自动映射到结构体字段。
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config 是应用根配置结构体，聚合所有子模块配置。
// `mapstructure` tag 指定 YAML/环境变量到结构体字段的映射键。
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
}

// ServerConfig HTTP 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"` // 监听端口，默认 8080
	Mode string `mapstructure:"mode"` // gin 模式：debug / release
}

// DatabaseConfig PostgreSQL 连接配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
	Timezone string `mapstructure:"timezone"`
}

// RedisConfig Redis 连接配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// JWTConfig JWT 签名与过期配置
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`       // HMAC 签名密钥（生产环境务必更换）
	ExpireHours int    `mapstructure:"expire_hours"` // Token 有效期（小时）
	Issuer      string `mapstructure:"issuer"`       // 签发者标识
}

// RabbitMQConfig RabbitMQ 连接配置
type RabbitMQConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	VHost    string `mapstructure:"vhost"`
}

// GlobalConfig 全局配置单例，在 Load() 后可供全项目读取
var GlobalConfig *Config

// Default 返回一份开发环境的默认配置。
// 用于配置文件缺失或加载失败时的降级兜底。
func Default() *Config {
	return &Config{
		Server: ServerConfig{Port: 8080, Mode: "debug"},
		Database: DatabaseConfig{
			Host: "localhost", Port: 5432, User: "postgres", Password: "postgres",
			DBName: "casbin_demo", SSLMode: "disable", Timezone: "Asia/Shanghai",
		},
		Redis:    RedisConfig{Host: "localhost", Port: 6379, DB: 0, PoolSize: 10},
		JWT:      JWTConfig{Secret: "casbin-demo-jwt-secret-key-change-in-production", ExpireHours: 24, Issuer: "casbin-demo"},
		RabbitMQ: RabbitMQConfig{Host: "localhost", Port: 5672, User: "guest", Password: "guest", VHost: "/"},
	}
}

// Load 从指定路径加载 YAML 配置文件，并用环境变量覆盖同名键。
// 环境变量命名规则：将配置键的 . 替换为 _，全大写。
//
// 例如：
//
//	database.host → DATABASE_HOST
//	redis.port    → REDIS_PORT
//
// configPath 为空时将直接使用默认配置。
func Load(configPath string) (*Config, error) {
	v := viper.New()

	if configPath != "" {
		v.SetConfigFile(configPath)
		v.SetConfigType("yaml")
		if err := v.ReadInConfig(); err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("read config file: %w", err)
			}
			slog.Warn("config file not found, falling back to env/defaults", "path", configPath)
		}
	}

	// 自动将 NESTED_KEY 形式的环境变量绑定到 nested.key 配置键
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 显式绑定常用环境变量，避免 viper 自动绑定不生效
	bindEnvs(v,
		"database.host", "database.port", "database.user", "database.password",
		"database.dbname", "database.sslmode", "database.timezone",
		"redis.host", "redis.port", "redis.password", "redis.db", "redis.pool_size",
		"server.port", "server.mode",
		"jwt.secret", "jwt.expire_hours", "jwt.issuer",
		"rabbitmq.host", "rabbitmq.port", "rabbitmq.user", "rabbitmq.password", "rabbitmq.vhost",
	)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// 对空字段填充默认值，保证下游使用零值问题
	applyDefaults(&cfg)

	GlobalConfig = &cfg
	slog.Info("config loaded",
		"db", fmt.Sprintf("%s:%d", cfg.Database.Host, cfg.Database.Port),
		"redis", fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		"port", cfg.Server.Port,
		"mode", cfg.Server.Mode,
	)
	return &cfg, nil
}

// bindEnvs 批量绑定环境变量，忽略 BindEnv 自身的错误。
func bindEnvs(v *viper.Viper, keys ...string) {
	for _, k := range keys {
		_ = v.BindEnv(k)
	}
}

// applyDefaults 对 cfg 中的零值字段填充默认值。
// Go 1.22+ for range int 可简化循环，但此处仅做简单字段判断。
func applyDefaults(cfg *Config) {
	if cfg.Database.Host == "" {
		cfg.Database.Host = "localhost"
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 5432
	}
	if cfg.Database.User == "" {
		cfg.Database.User = "postgres"
	}
	if cfg.Database.Password == "" {
		cfg.Database.Password = "postgres"
	}
	if cfg.Database.DBName == "" {
		cfg.Database.DBName = "casbin_demo"
	}
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}
	if cfg.Database.Timezone == "" {
		cfg.Database.Timezone = "Asia/Shanghai"
	}
	if cfg.Redis.Host == "" {
		cfg.Redis.Host = "localhost"
	}
	if cfg.Redis.Port == 0 {
		cfg.Redis.Port = 6379
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "casbin-demo-jwt-secret-key-change-in-production"
	}
	if cfg.JWT.ExpireHours == 0 {
		cfg.JWT.ExpireHours = 24
	}
	if cfg.JWT.Issuer == "" {
		cfg.JWT.Issuer = "casbin-demo"
	}
	if cfg.RabbitMQ.Host == "" {
		cfg.RabbitMQ.Host = "localhost"
	}
	if cfg.RabbitMQ.Port == 0 {
		cfg.RabbitMQ.Port = 5672
	}
	if cfg.RabbitMQ.User == "" {
		cfg.RabbitMQ.User = "guest"
	}
	if cfg.RabbitMQ.Password == "" {
		cfg.RabbitMQ.Password = "guest"
	}
	if cfg.RabbitMQ.VHost == "" {
		cfg.RabbitMQ.VHost = "/"
	}
}

// DSN 返回 PostgreSQL 连接字符串（lib/pq 与 pgx 通用格式）。
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode, d.Timezone,
	)
}

// Addr 返回 Redis host:port 地址字符串。
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}
