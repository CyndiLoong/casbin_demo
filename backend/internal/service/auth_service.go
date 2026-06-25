// Package service 实现业务逻辑层（Business Logic Layer）。
//
// 本层职责：
//   - 封装业务规则和流程编排
//   - 调用 Repository 层进行数据访问
//   - 通过 cache.Client 实现读写缓存，内置缓存穿透/击穿/雪崩防护
//   - 处理密码哈希、JWT生成、权限策略同步等
//   - 写操作（CUD）完成后主动失效相关缓存
//   - 向上层（Handler）返回业务友好的错误信息
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
	jwtpkg "casbin-demo/pkg/jwt"
)

// AuthService 认证服务，处理登录、注册、用户信息获取。
//
// 缓存策略：
//   - 登录操作：每次都查数据库（安全要求：密码校验不应走缓存）
//   - GetUserInfo：缓存 UserResponse DTO（不含密码），TTL=30min+抖动
//   - 注册后失效用户相关缓存
type AuthService struct {
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
	cache    *cache.Client
}

// NewAuthService 创建 AuthService 实例。
func NewAuthService(userRepo *repository.UserRepository, roleRepo *repository.RoleRepository, cacheClient *cache.Client) *AuthService {
	return &AuthService{userRepo: userRepo, roleRepo: roleRepo, cache: cacheClient}
}

// toUserResponse 将 User 实体转换为脱敏 DTO。
func toUserResponse(user *model.User) model.UserResponse {
	roles := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
	}
	return model.UserResponse{
		ID:        user.ID,
		UUID:      user.UUID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Status:    user.Status,
		Roles:     roles,
		CreatedAt: user.CreatedAt,
	}
}

// Login 用户登录逻辑（不走缓存，每次查库验证密码）。
//
// 安全设计：
//   - 用户名不存在和密码错误返回相同错误信息，防止用户名枚举攻击
//   - 登录操作不缓存认证结果，密码校验必须走数据库
//   - 登录成功后，将 UserResponse 写入缓存供后续 GetUserInfo 使用
func (s *AuthService) Login(req model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, errors.New("用户不存在或密码错误")
	}

	if user.Status != 1 {
		return nil, errors.New("用户已被禁用")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		slog.Warn("login failed: invalid password", "username", req.Username)
		return nil, errors.New("用户不存在或密码错误")
	}

	roles := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
	}

	token, err := jwtpkg.GenerateToken(user.ID, user.UUID, user.Username, roles)
	if err != nil {
		return nil, errors.New("生成token失败")
	}

	resp := toUserResponse(user)
	s.cache.SetLogical(cache.CacheKeyUint("user", "id", user.ID), &resp, cache.TTLHot)
	s.cache.BloomFilter().Add(cache.CacheKeyUint("user", "id", user.ID))

	slog.Info("user login success", "username", user.Username, "roles", roles)

	return &model.LoginResponse{Token: token, User: resp}, nil
}

// Register 用户注册逻辑。
//
// 处理流程：
//  1. 检查用户名是否已存在
//  2. 使用 bcrypt 加密密码（默认成本 10）
//  3. 生成 UUID、创建用户记录
//  4. 自动分配 "user" 角色
//  5. 失效用户相关缓存
func (s *AuthService) Register(req model.RegisterRequest) (*model.User, error) {
	_, err := s.userRepo.FindByUsername(req.Username)
	if err == nil {
		return nil, errors.New("用户名已存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}

	user := &model.User{
		UUID:      uuid.New().String(),
		Username:  req.Username,
		Password:  string(hashedPassword),
		Nickname:  nickname,
		Email:     req.Email,
		Status:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("创建用户失败")
	}

	userRole, err := s.roleRepo.FindByName("user")
	if err == nil {
		_ = s.userRepo.AssignRole(user.ID, userRole.ID)
		slog.Info("new user registered", "username", user.Username, "role", "user")
	}

	s.cache.InvalidateUserCaches()
	return user, nil
}

// GetUserInfo 获取用户信息（走热数据缓存，逻辑过期+异步重建防击穿）。
//
// 缓存策略：热点数据，使用逻辑过期（永不过期+异步重建），彻底杜绝缓存击穿。
// 布隆过滤器预热合法用户ID，防止缓存穿透。
func (s *AuthService) GetUserInfo(userID uint) (*model.UserResponse, error) {
	key := cache.CacheKeyUint("user", "id", userID)
	var cached model.UserResponse
	opt := cache.HotDataOptions()
	v, found, err := s.cache.Fetch(key, opt, &cached, func() (interface{}, error) {
		slog.Debug("cache miss: load user from db", "user_id", userID)
		user, e := s.userRepo.FindByID(userID)
		if e != nil {
			return nil, nil
		}
		resp := toUserResponse(user)
		return &resp, nil
	})
	if err != nil {
		return nil, errors.New("获取用户信息失败")
	}
	if !found || v == nil {
		return nil, errors.New("用户不存在")
	}
	if resp, ok := v.(*model.UserResponse); ok {
		return resp, nil
	}
	return nil, errors.New("用户不存在")
}
