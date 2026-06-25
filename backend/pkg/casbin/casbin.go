// Package casbin 封装 Casbin RBAC 权限引擎的初始化和策略同步逻辑。
//
// Casbin 模型：RBAC with domains（本项目不使用 domain）
// 策略格式：p = sub, obj, act （用户/角色, 资源路径, HTTP方法）
// 角色继承：g = _, _ （用户 → 角色）
//
// 匹配函数：
//   - keyMatch2: 支持 /api/users/:id 这种路径参数匹配
package casbin

import (
	"log/slog"
	"os"
	"path/filepath"

	casbin "github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// Enforcer 全局 Casbin 权限执行器实例。
// 在 InitCasbin() 成功后初始化，供中间件和服务层使用。
var Enforcer *casbin.Enforcer

// findModelPath 查找 Casbin 模型配置文件 model.conf 的路径。
//
// 按候选路径依次查找，支持从项目根目录或 backend 子目录运行：
//   - pkg/casbin/model.conf （从 backend 目录运行时）
//   - backend/pkg/casbin/model.conf （从项目根目录运行时）
//
// 找不到时返回第一个候选路径作为兜底（后续初始化会报错）。
func findModelPath() string {
	candidates := []string{
		"pkg/casbin/model.conf",
		"backend/pkg/casbin/model.conf",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			absPath, _ := filepath.Abs(p)
			slog.Debug("casbin model.conf found", "path", absPath)
			return p
		}
	}
	slog.Warn("casbin model.conf not found in standard paths, using fallback", "tried", candidates)
	return "pkg/casbin/model.conf"
}

// InitCasbin 初始化 Casbin RBAC 权限引擎。
//
// 使用 gorm-adapter 将策略持久化到 PostgreSQL，自动加载数据库中已有的策略。
// 初始化成功后将 Enforcer 赋值给包级变量供全局使用。
//
// 参数：
//   - db: GORM 数据库实例（用于 gorm-adapter 策略持久化）
//
// 返回：初始化后的 Enforcer 实例和错误。
func InitCasbin(db *gorm.DB) (*casbin.Enforcer, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, err
	}

	modelPath := findModelPath()
	enforcer, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		return nil, err
	}

	if err := enforcer.LoadPolicy(); err != nil {
		return nil, err
	}

	Enforcer = enforcer
	slog.Info("casbin enforcer initialized", "model", "RBAC", "adapter", "gorm", "model_path", modelPath)
	return enforcer, nil
}

// LoadCasbinPolicy 从数据库同步 Casbin 策略。
//
// 同步策略：清空内存中的策略，然后从 roles/permissions/user_roles 表重新加载。
// 在角色/权限变更（增删改、分配权限）后调用此函数保持 Casbin 与数据库一致。
//
// 加载内容：
//  1. p 策略：角色 → (路径, 方法) 的访问权限
//  2. g 策略：用户 → 角色 的归属关系
func LoadCasbinPolicy(e *casbin.Enforcer, db *gorm.DB) error {
	// 查询所有角色
	roles := []struct {
		Name string
	}{}
	if err := db.Table("roles").Select("name").Find(&roles).Error; err != nil {
		return err
	}

	// 查询角色-权限关联（角色 → 路径+方法）
	permissions := []struct {
		RoleName string
		Path     string
		Method   string
	}{}
	if err := db.Raw(`
		SELECT r.name as role_name, p.path, p.method
		FROM roles r
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
	`).Scan(&permissions).Error; err != nil {
		return err
	}

	e.ClearPolicy()

	for _, p := range permissions {
		if _, err := e.AddPolicy(p.RoleName, p.Path, p.Method); err != nil {
			slog.Warn("casbin add policy failed", "role", p.RoleName, "path", p.Path, "method", p.Method, "error", err)
		}
	}

	// 查询用户-角色关联（用户 → 角色）
	userRoles := []struct {
		Username string
		RoleName string
	}{}
	if err := db.Raw(`
		SELECT u.username, r.name as role_name
		FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		JOIN roles r ON ur.role_id = r.id
	`).Scan(&userRoles).Error; err != nil {
		return err
	}

	for _, ur := range userRoles {
		if _, err := e.AddGroupingPolicy(ur.Username, ur.RoleName); err != nil {
			slog.Warn("casbin add grouping policy failed", "user", ur.Username, "role", ur.RoleName, "error", err)
		}
	}

	if err := e.SavePolicy(); err != nil {
		return err
	}

	slog.Info("casbin policy reloaded",
		"policies", len(permissions),
		"grouping_policies", len(userRoles),
	)
	return nil
}
