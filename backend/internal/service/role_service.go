// Package service 实现业务逻辑层（Business Logic Layer）。
package service

import (
	"errors"
	"log/slog"

	"casbin-demo/internal/model"
	"casbin-demo/internal/repository"
	"casbin-demo/pkg/cache"
	casbinpkg "casbin-demo/pkg/casbin"
)

// RoleService 角色管理服务，处理角色 CRUD 和权限分配。
//
// 缓存策略：
//   - List（全量角色列表）：配置类缓存，TTL=1h+抖动（角色数据量小、变更频率低）
//   - 所有写操作（Create/Update/Delete/AssignPermission）：
//     a. 失效角色相关缓存
//     b. 失效权限列表缓存（角色关联权限）
//     c. 失效用户缓存（角色变化影响用户的 roles 字段）
//     d. 失效仪表盘统计缓存
//     e. 同步 Casbin 策略使权限即时生效
type RoleService struct {
	roleRepo *repository.RoleRepository
	cache    *cache.Client
}

// NewRoleService 创建 RoleService 实例。
func NewRoleService(roleRepo *repository.RoleRepository, cacheClient *cache.Client) *RoleService {
	return &RoleService{roleRepo: roleRepo, cache: cacheClient}
}

// List 查询所有角色（走配置类缓存）。
//
// 缓存策略：配置类缓存 TTL=1h+抖动，角色数据量小且变更频率低，适合长 TTL。
// 缓存 Key: cache:role:list
//
// 注意：角色列表包含关联的 Permissions 数据，缓存的是完整的 []model.Role。
func (s *RoleService) List() ([]model.Role, error) {
	key := cache.CacheKey("role", "list", "all")
	var cached []model.Role
	opt := cache.DefaultFetchOptions(cache.TTLConfig)
	opt.UseLogicalExp = true
	v, found, err := s.cache.Fetch(key, opt, &cached, func() (interface{}, error) {
		slog.Debug("cache miss: load role list from db")
		return s.roleRepo.List()
	})
	if err != nil {
		return nil, errors.New("获取角色列表失败")
	}
	if !found || v == nil {
		return nil, errors.New("获取角色列表失败")
	}
	if roles, ok := v.(*[]model.Role); ok {
		return *roles, nil
	}
	return cached, nil
}

// Create 创建新角色。
//
// 校验角色名唯一性后创建。
// 创建后不立即同步 Casbin（因为角色尚未分配权限），但失效相关缓存。
func (s *RoleService) Create(req model.CreateRoleRequest) (*model.Role, error) {
	_, err := s.roleRepo.FindByName(req.Name)
	if err == nil {
		return nil, errors.New("角色名称已存在")
	}

	status := req.Status
	if status == 0 {
		status = 1
	}

	role := &model.Role{
		Name:        req.Name,
		Label:       req.Label,
		Description: req.Description,
		Status:      status,
	}

	if err := s.roleRepo.Create(role); err != nil {
		return nil, errors.New("创建角色失败")
	}

	s.cache.InvalidateRoleCaches()
	slog.Info("role created", "name", req.Name)
	return role, nil
}

// Update 更新角色信息。
//
// 更新成功后：
//  1. 失效角色相关缓存
//  2. 重新加载 Casbin 策略（角色名变更可能影响策略匹配）
func (s *RoleService) Update(id uint, req model.CreateRoleRequest) error {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		return errors.New("角色不存在")
	}

	role.Name = req.Name
	role.Label = req.Label
	role.Description = req.Description
	if req.Status != 0 {
		role.Status = req.Status
	}

	if err := s.roleRepo.Update(role); err != nil {
		return errors.New("更新角色失败")
	}

	s.cache.InvalidateRoleCaches()
	slog.Info("role updated", "role_id", id)

	if casbinpkg.Enforcer != nil {
		_ = casbinpkg.LoadCasbinPolicy(casbinpkg.Enforcer, repository.DB)
	}

	return nil
}

// Delete 删除角色（软删除）。
//
// 删除成功后失效角色相关缓存。
func (s *RoleService) Delete(id uint) error {
	if err := s.roleRepo.Delete(id); err != nil {
		return errors.New("删除角色失败")
	}

	s.cache.InvalidateRoleCaches()
	slog.Info("role deleted", "role_id", id)
	return nil
}

// AssignPermission 为角色分配权限。
//
// 分配成功后：
//  1. 失效角色和权限相关缓存
//  2. 重新加载 Casbin 策略使权限立即生效
func (s *RoleService) AssignPermission(roleID, permissionID uint) error {
	if err := s.roleRepo.AssignPermission(roleID, permissionID); err != nil {
		return errors.New("分配权限失败")
	}

	s.cache.InvalidateRoleCaches()
	slog.Info("permission assigned to role", "role_id", roleID, "permission_id", permissionID)

	if casbinpkg.Enforcer != nil {
		_ = casbinpkg.LoadCasbinPolicy(casbinpkg.Enforcer, repository.DB)
	}

	return nil
}
