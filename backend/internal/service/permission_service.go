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

// PermissionService 权限管理服务，处理权限 CRUD。
//
// 缓存策略：
//   - List（全量权限列表）：配置类缓存，TTL=1h+抖动（权限数据量小、变更频率极低）
//   - 所有写操作（Create/Update/Delete）：
//     a. 失效权限相关缓存
//     b. 失效角色列表缓存（角色关联权限）
//     c. 失效仪表盘统计缓存
//     d. 同步 Casbin 策略使权限即时生效
type PermissionService struct {
	permRepo *repository.PermissionRepository
	cache    *cache.Client
}

// NewPermissionService 创建 PermissionService 实例。
func NewPermissionService(permRepo *repository.PermissionRepository, cacheClient *cache.Client) *PermissionService {
	return &PermissionService{permRepo: permRepo, cache: cacheClient}
}

// List 查询所有权限（走配置类缓存）。
//
// 缓存策略：配置类缓存 TTL=1h+抖动，权限是系统配置数据，变更频率极低。
// 缓存 Key: cache:permission:list
func (s *PermissionService) List() ([]model.Permission, error) {
	key := cache.CacheKey("permission", "list", "all")
	var cached []model.Permission
	opt := cache.DefaultFetchOptions(cache.TTLConfig)
	opt.UseLogicalExp = true
	v, found, err := s.cache.Fetch(key, opt, &cached, func() (interface{}, error) {
		slog.Debug("cache miss: load permission list from db")
		return s.permRepo.List()
	})
	if err != nil {
		return nil, errors.New("获取权限列表失败")
	}
	if !found || v == nil {
		return nil, errors.New("获取权限列表失败")
	}
	if perms, ok := v.(*[]model.Permission); ok {
		return *perms, nil
	}
	return cached, nil
}

// Create 创建新权限。
//
// 创建成功后：
//  1. 失效权限相关缓存
//  2. 重新加载 Casbin 策略使新权限生效
func (s *PermissionService) Create(req model.CreatePermissionRequest) (*model.Permission, error) {
	perm := &model.Permission{
		Name:        req.Name,
		Label:       req.Label,
		Description: req.Description,
		Path:        req.Path,
		Method:      req.Method,
	}

	if err := s.permRepo.Create(perm); err != nil {
		return nil, errors.New("创建权限失败")
	}

	s.cache.InvalidatePermissionCaches()
	slog.Info("permission created", "name", req.Name, "path", req.Path, "method", req.Method)

	if casbinpkg.Enforcer != nil {
		_ = casbinpkg.LoadCasbinPolicy(casbinpkg.Enforcer, repository.DB)
	}

	return perm, nil
}

// Update 更新权限信息。
//
// 更新成功后：
//  1. 失效权限相关缓存
//  2. 重新加载 Casbin 策略（路径/方法变更会影响策略匹配）
func (s *PermissionService) Update(id uint, req model.CreatePermissionRequest) error {
	perm, err := s.permRepo.FindByID(id)
	if err != nil {
		return errors.New("权限不存在")
	}

	perm.Name = req.Name
	perm.Label = req.Label
	perm.Description = req.Description
	perm.Path = req.Path
	perm.Method = req.Method

	if err := s.permRepo.Update(perm); err != nil {
		return errors.New("更新权限失败")
	}

	s.cache.InvalidatePermissionCaches()
	slog.Info("permission updated", "permission_id", id)

	if casbinpkg.Enforcer != nil {
		_ = casbinpkg.LoadCasbinPolicy(casbinpkg.Enforcer, repository.DB)
	}

	return nil
}

// Delete 删除权限（软删除）。
//
// 删除成功后：
//  1. 失效权限相关缓存
//  2. 重新加载 Casbin 策略移除已删除的权限
func (s *PermissionService) Delete(id uint) error {
	if err := s.permRepo.Delete(id); err != nil {
		return errors.New("删除权限失败")
	}

	s.cache.InvalidatePermissionCaches()
	slog.Info("permission deleted", "permission_id", id)

	if casbinpkg.Enforcer != nil {
		_ = casbinpkg.LoadCasbinPolicy(casbinpkg.Enforcer, repository.DB)
	}

	return nil
}
