package app

import (
	"log/slog"
	"os"

	"gorm.io/gorm"

	"casbin-demo/internal/config"
	"casbin-demo/internal/repository"
)

// setupSlog 配置结构化日志（JSON 格式输出到 stdout）。
func setupSlog() {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: false,
	})
	slog.SetDefault(slog.New(h))
}

// mustLoadConfig 加载 YAML 配置文件，失败时降级到默认配置。
func mustLoadConfig() *config.Config {
	appCfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Warn("config load warning, using defaults", "error", err)
		appCfg = config.Default()
	}
	return appCfg
}

// mustInitDB 初始化 PostgreSQL 数据库连接，失败时 panic（由 fx 捕获）。
func mustInitDB(appCfg *config.Config) *gorm.DB {
	db, err := repository.InitDB(appCfg)
	if err != nil {
		slog.Error("database initialization failed", "error", err)
		panic(err)
	}
	slog.Info("postgresql connected", "host", appCfg.Database.Host, "port", appCfg.Database.Port)
	return db
}

// mustSeedData 幂等初始化种子数据，失败时 panic（由 fx 捕获）。
func mustSeedData(db *gorm.DB) {
	if err := repository.SeedData(db); err != nil {
		slog.Error("seed data failed", "error", err)
		panic(err)
	}
	slog.Info("seed data initialized")
}
