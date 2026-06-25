// Package repository 实现数据访问层（Data Access Layer）。
//
// 本层职责：
//   - 封装数据库（PostgreSQL）和缓存（Redis）的连接初始化
//   - 提供各实体的 CRUD 操作 Repository
//   - 管理数据库连接池、自动迁移（AutoMigrate）、种子数据
//   - 向上层（Service）屏蔽底层存储细节
//
// 设计原则：
//   - 依赖 GORM 实现 ORM 映射
//   - 使用 %w 错误包装保留错误链
//   - 种子数据使用幂等写法，重复执行不会报错
package repository

import (
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"casbin-demo/internal/config"
	"casbin-demo/internal/model"
)

// DB 全局 GORM DB 实例，供各 Repository 使用。
// 在 InitDB() 成功后初始化。
var DB *gorm.DB

// InitDB 初始化 PostgreSQL 数据库连接。
//
// 执行步骤：
//  1. 使用配置 DSN 打开 PostgreSQL 连接
//  2. 获取底层 sql.DB 设置连接池参数
//     - MaxIdleConns: 10（空闲连接数）
//     - MaxOpenConns: 100（最大打开连接数）
//  3. 执行 AutoMigrate 自动创建/更新表结构
//  4. 保存全局 DB 实例
//
// 注意：生产环境应使用 SQL 迁移工具（如 golang-migrate）替代 AutoMigrate。
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	if err := db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
	); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	slog.Info("postgresql connected and migrated",
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
		"dbname", cfg.Database.DBName,
	)
	DB = db
	return db, nil
}

// SeedData 初始化种子数据（幂等操作）。
//
// 初始化内容：
//  1. 默认角色：admin（管理员）、user（普通用户）
//  2. 默认权限：dashboard/user/role/permission 系列接口
//  3. admin 角色授予所有权限
//  4. 默认用户：admin/admin123（管理员）、user/user123（普通用户）
//
// 密码均为 bcrypt 哈希值，对应明文：admin123 / user123
func SeedData(db *gorm.DB) error {
	var adminRole model.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		adminRole = model.Role{
			Name:        "admin",
			Label:       "管理员",
			Description: "系统管理员，拥有所有权限",
			Status:      1,
		}
		if err := db.Create(&adminRole).Error; err != nil {
			return fmt.Errorf("create admin role: %w", err)
		}
		slog.Info("seed: created admin role")
	}

	var userRole model.Role
	if err := db.Where("name = ?", "user").First(&userRole).Error; err != nil {
		userRole = model.Role{
			Name:        "user",
			Label:       "普通用户",
			Description: "普通用户，拥有基础权限",
			Status:      1,
		}
		if err := db.Create(&userRole).Error; err != nil {
			return fmt.Errorf("create user role: %w", err)
		}
		slog.Info("seed: created user role")
	}

	permissions := []model.Permission{
		{Name: "dashboard", Label: "仪表盘", Path: "/api/dashboard", Method: "GET", Description: "查看仪表盘"},
		{Name: "user:list", Label: "用户列表", Path: "/api/users", Method: "GET", Description: "查看用户列表"},
		{Name: "user:create", Label: "创建用户", Path: "/api/users", Method: "POST", Description: "创建用户"},
		{Name: "user:update", Label: "更新用户", Path: "/api/users/:id", Method: "PUT", Description: "更新用户"},
		{Name: "user:delete", Label: "删除用户", Path: "/api/users/:id", Method: "DELETE", Description: "删除用户"},
		{Name: "role:list", Label: "角色列表", Path: "/api/roles", Method: "GET", Description: "查看角色列表"},
		{Name: "role:create", Label: "创建角色", Path: "/api/roles", Method: "POST", Description: "创建角色"},
		{Name: "role:update", Label: "更新角色", Path: "/api/roles/:id", Method: "PUT", Description: "更新角色"},
		{Name: "role:delete", Label: "删除角色", Path: "/api/roles/:id", Method: "DELETE", Description: "删除角色"},
		{Name: "permission:list", Label: "权限列表", Path: "/api/permissions", Method: "GET", Description: "查看权限列表"},
	}

	for _, p := range permissions {
		var perm model.Permission
		if err := db.Where("name = ?", p.Name).First(&perm).Error; err != nil {
			if err := db.Create(&p).Error; err != nil {
				slog.Warn("create permission failed", "name", p.Name, "error", err)
			}
		}
	}

	var allPerms []model.Permission
	db.Find(&allPerms)
	if err := db.Model(&adminRole).Association("Permissions").Replace(allPerms); err != nil {
		slog.Warn("assign permissions to admin failed", "error", err)
	} else {
		slog.Info("seed: assigned all permissions to admin role", "count", len(allPerms))
	}

	var adminUser model.User
	if err := db.Where("username = ?", "admin").First(&adminUser).Error; err != nil {
		adminPasswordHash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		adminUser = model.User{
			UUID:      "00000000-0000-0000-0000-000000000001",
			Username:  "admin",
			Password:  string(adminPasswordHash),
			Nickname:  "管理员",
			Email:     "admin@example.com",
			Status:    1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := db.Create(&adminUser).Error; err != nil {
			return fmt.Errorf("create admin user: %w", err)
		}
		if err := db.Model(&adminUser).Association("Roles").Append(&adminRole); err != nil {
			return fmt.Errorf("assign admin role: %w", err)
		}
		slog.Info("seed: created admin user (password: admin123)")
	}

	var testUser model.User
	if err := db.Where("username = ?", "user").First(&testUser).Error; err != nil {
		userPasswordHash, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
		testUser = model.User{
			UUID:      "00000000-0000-0000-0000-000000000002",
			Username:  "user",
			Password:  string(userPasswordHash),
			Nickname:  "普通用户",
			Email:     "user@example.com",
			Status:    1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := db.Create(&testUser).Error; err != nil {
			return fmt.Errorf("create test user: %w", err)
		}
		if err := db.Model(&testUser).Association("Roles").Append(&userRole); err != nil {
			return fmt.Errorf("assign user role: %w", err)
		}
		slog.Info("seed: created test user (password: user123)")
	}

	slog.Info("seed data initialized successfully")
	return nil
}
