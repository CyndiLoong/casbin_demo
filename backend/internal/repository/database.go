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
		&model.Resource{},
		&model.AuditApplication{},
		&model.SysMessage{},
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
		{Name: "user:assign-role", Label: "分配角色", Path: "/api/users/assign-role", Method: "POST", Description: "为用户分配角色"},
		{Name: "role:list", Label: "角色列表", Path: "/api/roles", Method: "GET", Description: "查看角色列表"},
		{Name: "role:create", Label: "创建角色", Path: "/api/roles", Method: "POST", Description: "创建角色"},
		{Name: "role:update", Label: "更新角色", Path: "/api/roles/:id", Method: "PUT", Description: "更新角色"},
		{Name: "role:delete", Label: "删除角色", Path: "/api/roles/:id", Method: "DELETE", Description: "删除角色"},
		{Name: "role:assign-permission", Label: "分配权限", Path: "/api/roles/assign-permission", Method: "POST", Description: "为角色分配权限"},
		{Name: "permission:list", Label: "权限列表", Path: "/api/permissions", Method: "GET", Description: "查看权限列表"},
		{Name: "permission:create", Label: "创建权限", Path: "/api/permissions", Method: "POST", Description: "创建权限"},
		{Name: "permission:update", Label: "更新权限", Path: "/api/permissions/:id", Method: "PUT", Description: "更新权限"},
		{Name: "permission:delete", Label: "删除权限", Path: "/api/permissions/:id", Method: "DELETE", Description: "删除权限"},
		{Name: "audit:list-all", Label: "审核列表(全部)", Path: "/api/audit/applications", Method: "GET", Description: "查看所有审核申请"},
		{Name: "audit:review", Label: "审核操作", Path: "/api/audit/applications/:id/review", Method: "POST", Description: "审核申请"},
		{Name: "audit:pending-count", Label: "待审核数量", Path: "/api/audit/pending-count", Method: "GET", Description: "获取待审核数量"},
		{Name: "resource:create", Label: "创建资源", Path: "/api/resources", Method: "POST", Description: "创建资源"},
		{Name: "resource:update", Label: "更新资源", Path: "/api/resources/:id", Method: "PUT", Description: "更新资源"},
		{Name: "resource:delete", Label: "删除资源", Path: "/api/resources/:id", Method: "DELETE", Description: "删除资源"},
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

	var userPerms []model.Permission
	db.Where("name IN ?", []string{"dashboard"}).Find(&userPerms)
	if err := db.Model(&userRole).Association("Permissions").Replace(userPerms); err != nil {
		slog.Warn("assign permissions to user failed", "error", err)
	} else {
		slog.Info("seed: assigned basic permissions to user role", "count", len(userPerms))
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

	// 初始化资源清单种子数据（幂等）
	seedResources(db)

	slog.Info("seed data initialized successfully")
	return nil
}

// seedResources 初始化资源清单种子数据（幂等操作）。
//
// 初始化内容：
//  - 预置常见的大模型API资源，包括对话大模型、代码大模型、图像生成、语音识别等
//  - 每个资源包含名称、类型、API名称、描述、厂商、版本、QPS配额等信息
//  - 使用 API名称 作为唯一标识，重复执行不会报错
func seedResources(db *gorm.DB) {
	resources := []model.Resource{
		{
			Name:         "GPT-4 对话大模型",
			Type:         "llm_chat",
			APIName:      "gpt-4-chat",
			Description:  "基于GPT-4的通用对话大模型，支持多轮对话、文本生成、逻辑推理等复杂任务",
			Provider:     "OpenAI",
			Version:      "gpt-4-0613",
			DefaultQPS:   10,
			MaxQPS:       100,
			Status:       model.ResourceStatusActive,
			DocsURL:      "https://platform.openai.com/docs/api-reference/chat",
			Tags:         `["大模型","对话","推理"]`,
		},
		{
			Name:         "GPT-3.5 对话大模型",
			Type:         "llm_chat",
			APIName:      "gpt-3.5-turbo",
			Description:  "基于GPT-3.5 Turbo的快速对话大模型，性价比高，适合日常对话场景",
			Provider:     "OpenAI",
			Version:      "gpt-3.5-turbo-0613",
			DefaultQPS:   20,
			MaxQPS:       200,
			Status:       model.ResourceStatusActive,
			DocsURL:      "https://platform.openai.com/docs/api-reference/chat",
			Tags:         `["大模型","对话","高性价比"]`,
		},
		{
			Name:         "GPT-4 代码大模型",
			Type:         "llm_code",
			APIName:      "gpt-4-code",
			Description:  "专门优化的代码生成和理解模型，支持代码补全、代码解释、Bug修复等",
			Provider:     "OpenAI",
			Version:      "gpt-4-0613",
			DefaultQPS:   5,
			MaxQPS:       50,
			Status:       model.ResourceStatusActive,
			DocsURL:      "https://platform.openai.com/docs/api-reference/chat",
			Tags:         `["大模型","代码生成","编程助手"]`,
		},
		{
			Name:         "DALL-E 3 图像生成",
			Type:         "image_gen",
			APIName:      "dall-e-3",
			Description:  "基于DALL-E 3的AI图像生成模型，支持文本描述生成高质量图像",
			Provider:     "OpenAI",
			Version:      "dall-e-3",
			DefaultQPS:   2,
			MaxQPS:       20,
			Status:       model.ResourceStatusActive,
			DocsURL:      "https://platform.openai.com/docs/api-reference/images",
			Tags:         `["图像生成","AIGC","创意设计"]`,
		},
		{
			Name:         "Whisper 语音识别",
			Type:         "asr",
			APIName:      "whisper",
			Description:  "基于Whisper的语音转文字服务，支持多种语言和方言，准确率高",
			Provider:     "OpenAI",
			Version:      "whisper-1",
			DefaultQPS:   10,
			MaxQPS:       100,
			Status:       model.ResourceStatusActive,
			DocsURL:      "https://platform.openai.com/docs/api-reference/audio",
			Tags:         `["语音识别","ASR","多语言"]`,
		},
		{
			Name:         "TTS 语音合成",
			Type:         "tts",
			APIName:      "tts-1",
			Description:  "高质量文本转语音服务，支持多种音色和语速调节",
			Provider:     "OpenAI",
			Version:      "tts-1",
			DefaultQPS:   10,
			MaxQPS:       100,
			Status:       model.ResourceStatusActive,
			DocsURL:      "https://platform.openai.com/docs/api-reference/audio",
			Tags:         `["语音合成","TTS","多音色"]`,
		},
		{
			Name:         "text-embedding 向量嵌入",
			Type:         "embedding",
			APIName:      "text-embedding-ada-002",
			Description:  "文本向量化服务，用于语义搜索、相似度计算、聚类分析等场景",
			Provider:     "OpenAI",
			Version:      "text-embedding-ada-002",
			DefaultQPS:   50,
			MaxQPS:       500,
			Status:       model.ResourceStatusActive,
			DocsURL:      "https://platform.openai.com/docs/api-reference/embeddings",
			Tags:         `["向量嵌入","Embedding","语义搜索"]`,
		},
		{
			Name:         "通义千问 对话大模型",
			Type:         "llm_chat",
			APIName:      "qwen-turbo",
			Description:  "阿里云通义千问大模型，支持中文理解和生成，响应速度快",
			Provider:     "阿里云",
			Version:      "qwen-turbo",
			DefaultQPS:   15,
			MaxQPS:       150,
			Status:       model.ResourceStatusActive,
			DocsURL:      "https://help.aliyun.com/document_detail/2400395.html",
			Tags:         `["大模型","对话","中文优化"]`,
		},
	}

	for _, r := range resources {
		var existing model.Resource
		if err := db.Where("api_name = ?", r.APIName).First(&existing).Error; err != nil {
			if err := db.Create(&r).Error; err != nil {
				slog.Warn("create resource failed", "api_name", r.APIName, "error", err)
			} else {
				slog.Info("seed: created resource", "name", r.Name)
			}
		}
	}
}
