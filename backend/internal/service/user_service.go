// Package service 实现业务逻辑层（Business Logic Layer）。
package service

import (
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"casbin-demo/internal/model"
	"casbin-demo/internal/repository"
	"casbin-demo/pkg/cache"
)

// userListCache 用户列表缓存结构体，用于分页列表的序列化/反序列化。
//
// 为什么不直接缓存 []model.UserResponse？
// 因为分页列表需要同时缓存 list 数据和 total 总数，定义结构体便于统一管理。
type userListCache struct {
	List  []model.UserResponse `json:"list"`
	Total int64                `json:"total"`
}

// UserService 用户管理服务，处理用户 CRUD 和角色分配。
//
// 缓存策略：
//   - List（分页列表）：查询缓存，TTL=5min+抖动
//   - 单条用户详情：由 AuthService.GetUserInfo 缓存（热数据 TTL=30min）
//   - 所有写操作（Create/Update/Delete/AssignRole）：主动失效用户相关缓存
//   - 写操作同时失效仪表盘统计缓存（用户总数变化）
type UserService struct {
	userRepo *repository.UserRepository
	cache    *cache.Client
}

// NewUserService 创建 UserService 实例。
func NewUserService(userRepo *repository.UserRepository, cacheClient *cache.Client) *UserService {
	return &UserService{userRepo: userRepo, cache: cacheClient}
}

// BuildUserListResp 将 User 实体列表转换为脱敏 DTO 列表（内部工具函数）。
func BuildUserListResp(users []model.User, total int64) *userListCache {
	result := make([]model.UserResponse, 0, len(users))
	for _, user := range users {
		roles := make([]string, 0, len(user.Roles))
		for _, role := range user.Roles {
			roles = append(roles, role.Name)
		}
		result = append(result, model.UserResponse{
			ID:        user.ID,
			UUID:      user.UUID,
			Username:  user.Username,
			Nickname:  user.Nickname,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Status:    user.Status,
			Roles:     roles,
			CreatedAt: user.CreatedAt,
		})
	}
	return &userListCache{List: result, Total: total}
}

// List 分页查询用户列表（走查询缓存）。
//
// 缓存策略：查询缓存 TTL=5min+抖动，按分页参数构建 Key。
// 缓存 Key: cache:user:list:p{page}:s{pageSize}
//
// 缓存流程（由 cache.Fetch 统一处理）：
//  1. 查 Redis，命中则直接返回
//  2. 未命中：singleflight 合并并发请求（防击穿）
//  3. loader 从数据库加载数据
//  4. 数据为空则写空值标记（防穿透），数据非空则写缓存（TTL+抖动防雪崩）
func (s *UserService) List(page, pageSize int) ([]model.UserResponse, int64, error) {
	key := cache.CacheKeyList("user", page, pageSize)
	var cached userListCache
	opt := cache.DefaultFetchOptions(cache.TTLQuery)
	v, found, err := s.cache.Fetch(key, opt, &cached, func() (interface{}, error) {
		slog.Debug("cache miss: load user list from db", "page", page, "page_size", pageSize)
		users, total, e := s.userRepo.List(page, pageSize)
		if e != nil {
			return nil, e
		}
		return BuildUserListResp(users, total), nil
	})
	if err != nil {
		return nil, 0, errors.New("获取用户列表失败")
	}
	if !found || v == nil {
		return nil, 0, errors.New("获取用户列表失败")
	}
	if resp, ok := v.(*userListCache); ok {
		return resp.List, resp.Total, nil
	}
	return cached.List, cached.Total, nil
}

// Create 创建新用户（管理员接口）。
//
// 与 Register 不同，此接口：
//   - 不需要自动分配 user 角色（可后续通过 AssignRole 分配）
//   - 不返回 JWT Token
//
// 缓存操作：创建成功后失效所有用户相关缓存。
func (s *UserService) Create(username, password, nickname, email string) (*model.User, error) {
	_, err := s.userRepo.FindByUsername(username)
	if err == nil {
		return nil, errors.New("用户名已存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	if nickname == "" {
		nickname = username
	}

	user := &model.User{
		UUID:      uuid.New().String(),
		Username:  username,
		Password:  string(hashedPassword),
		Nickname:  nickname,
		Email:     email,
		Status:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("创建用户失败")
	}

	s.cache.InvalidateUserCaches()
	slog.Info("user created by admin", "username", username)
	return user, nil
}

// Update 更新用户信息（昵称、邮箱、状态）。
//
// 缓存操作：更新成功后失效所有用户相关缓存（列表缓存 + 单条缓存）。
func (s *UserService) Update(id uint, nickname, email string, status int) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("用户不存在")
	}

	user.Nickname = nickname
	user.Email = email
	user.Status = status
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return errors.New("更新用户失败")
	}

	s.cache.InvalidateUserCaches()
	slog.Info("user updated", "user_id", id)
	return nil
}

// Delete 删除用户（软删除）。
//
// 缓存操作：删除成功后失效所有用户相关缓存。
func (s *UserService) Delete(id uint) error {
	if err := s.userRepo.Delete(id); err != nil {
		return errors.New("删除用户失败")
	}

	s.cache.InvalidateUserCaches()
	slog.Info("user deleted", "user_id", id)
	return nil
}

// AssignRole 为用户分配角色。
//
// 缓存操作：分配成功后失效用户相关缓存（角色变化影响用户信息中的 roles 字段）。
func (s *UserService) AssignRole(userID, roleID uint) error {
	if err := s.userRepo.AssignRole(userID, roleID); err != nil {
		return errors.New("分配角色失败")
	}

	s.cache.InvalidateUserCaches()
	slog.Info("role assigned to user", "user_id", userID, "role_id", roleID)
	return nil
}
