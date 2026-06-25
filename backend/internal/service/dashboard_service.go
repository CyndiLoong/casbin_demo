// Package service 实现业务逻辑层（Business Logic Layer）。
package service

import (
	"log/slog"

	"casbin-demo/internal/repository"
	"casbin-demo/pkg/cache"
)

// DashboardStats 仪表盘统计数据结构体。
type DashboardStats struct {
	TotalUsers       int64 `json:"total_users"`
	TotalRoles       int64 `json:"total_roles"`
	TotalPermissions int64 `json:"total_permissions"`
}

// DashboardService 仪表盘服务，处理统计数据查询。
//
// 缓存策略：
//   - GetStats：仪表盘统计缓存，TTL=2min+抖动
//   - 任何用户/角色/权限的写操作都会失效仪表盘缓存（通过 InvalidateDashboardCache）
//   - 统计数据为聚合查询，短 TTL 平衡数据新鲜度与数据库压力
type DashboardService struct {
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
	permRepo *repository.PermissionRepository
	cache    *cache.Client
}

// NewDashboardService 创建 DashboardService 实例。
func NewDashboardService(
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	permRepo *repository.PermissionRepository,
	cacheClient *cache.Client,
) *DashboardService {
	return &DashboardService{
		userRepo: userRepo,
		roleRepo: roleRepo,
		permRepo: permRepo,
		cache:    cacheClient,
	}
}

// GetStats 获取仪表盘统计数据（走缓存）。
//
// 缓存策略：仪表盘统计缓存 TTL=2min+抖动。
// 缓存 Key: cache:dashboard:stats
//
// 数据来源：分别从 users/roles/permissions 三张表 COUNT(*)。
// 任何相关写操作都会主动失效此缓存，保证统计数据不会过度陈旧。
func (s *DashboardService) GetStats() (*DashboardStats, error) {
	key := cache.CacheKey("dashboard", "stats", "all")
	var cached DashboardStats
	opt := cache.DefaultFetchOptions(cache.TTLDashboard)
	v, found, err := s.cache.Fetch(key, opt, &cached, func() (interface{}, error) {
		slog.Debug("cache miss: load dashboard stats from db")

		userCount, e := s.userRepo.Count()
		if e != nil {
			return nil, e
		}
		roleCount, e := s.roleRepo.Count()
		if e != nil {
			return nil, e
		}
		permCount, e := s.permRepo.Count()
		if e != nil {
			return nil, e
		}

		return &DashboardStats{
			TotalUsers:       userCount,
			TotalRoles:       roleCount,
			TotalPermissions: permCount,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	if !found || v == nil {
		return nil, nil
	}
	if stats, ok := v.(*DashboardStats); ok {
		return stats, nil
	}
	return &cached, nil
}
