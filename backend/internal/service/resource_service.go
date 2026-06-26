// Package service 实现业务逻辑层（Business Logic Layer）。
//
// 本文件处理资源清单相关的业务逻辑，包括资源的增删改查和缓存策略。
//
// 缓存策略：
//   - 资源列表：查询缓存，TTL=5min±10%抖动，防止缓存雪崩
//   - 资源详情：热点数据缓存，逻辑过期+异步重建，防止缓存击穿
//   - 写操作后主动失效相关缓存（SCAN+DEL 模式，避免阻塞 Redis）
//   - 缓存穿透：空值标记+60s TTL，防止恶意查询不存在的ID
package service

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"casbin-demo/internal/model"
	"casbin-demo/internal/repository"
	"casbin-demo/pkg/cache"
)

// ResourceService 资源服务，处理资源清单的查询和管理。
//
// 业务边界：
//   - 普通用户：可查看资源列表、资源详情
//   - 管理员：可创建、更新、删除资源
//
// 缓存策略：
//   - 资源列表：查询级缓存（TTL 5min±10%）
//   - 资源详情：热点数据缓存（逻辑过期，异步重建）
//   - 写操作：主动失效相关缓存 key
type ResourceService struct {
	resourceRepo *repository.ResourceRepository
	cache        *cache.Client
}

// NewResourceService 创建 ResourceService 实例。
//
// 参数：
//   - resourceRepo: 资源数据访问对象
//   - cacheClient: 缓存客户端（多级缓存）
func NewResourceService(resourceRepo *repository.ResourceRepository, cacheClient *cache.Client) *ResourceService {
	return &ResourceService{resourceRepo: resourceRepo, cache: cacheClient}
}

// buildResourceListCacheKey 构建资源列表缓存 key。
//
// 将所有筛选参数拼接到 key 中，确保不同筛选条件有独立缓存。
// 格式：cache:resource:list:p{page}:s{pageSize}:t{type}:st{status}:k{keyword}
func buildResourceListCacheKey(page, pageSize int, resourceType string, status *int, keyword string) string {
	statusStr := "all"
	if status != nil {
		statusStr = fmt.Sprintf("%d", *status)
	}
	return cache.CacheKey("resource", "list",
		fmt.Sprintf("p%d:s%d:t%s:st%s:k%s", page, pageSize, resourceType, statusStr, keyword))
}

// invalidateResourceListCaches 失效所有资源列表相关缓存。
//
// 使用 SCAN 命令匹配所有 cache:resource:list:* 模式的 key 并批量删除，
// 避免在写操作后需要知道所有可能的列表筛选参数组合。
// 注意：SCAN 采用游标迭代，不会阻塞 Redis。
func (s *ResourceService) invalidateResourceListCaches() {
	if s.cache == nil {
		return
	}
	pattern := cache.CacheKey("resource", "list", "*")
	s.cache.DeleteByPattern(pattern)
	// 同时失效可用资源列表缓存
	activeKey := cache.CacheKey("resource", "active_list", "all")
	s.cache.Delete(activeKey)
}

// GetResourceList 分页查询资源列表（走查询缓存）。
//
// 缓存策略：
//   - 缓存类型：查询级缓存
//   - TTL：5分钟 ± 10% 抖动（防雪崩）
//   - 缓存穿透：空结果标记 60s TTL
//   - 缓存击穿：singleflight + 分布式锁
//
// 参数：
//   - page: 页码（从 1 开始）
//   - pageSize: 每页数量
//   - resourceType: 资源类型筛选（空字符串不筛选）
//   - status: 状态筛选（nil 不筛选）
//   - keyword: 关键词搜索（空字符串不搜索）
//
// 返回：资源列表、总记录数、错误。
func (s *ResourceService) GetResourceList(page, pageSize int, resourceType string, status *int, keyword string) ([]model.Resource, int64, error) {
	cacheKey := buildResourceListCacheKey(page, pageSize, resourceType, status, keyword)

	type listResult struct {
		List  []model.Resource `json:"list"`
		Total int64            `json:"total"`
	}

	var result listResult

	opt := cache.DefaultFetchOptions(cache.TTLQuery)
	opt.UseBloom = false
	data, found, err := s.cache.Fetch(cacheKey, opt, &result, func() (interface{}, error) {
		slog.Debug("cache miss: load resource list from db",
			"page", page, "page_size", pageSize, "type", resourceType)
		list, total, e := s.resourceRepo.List(page, pageSize, resourceType, status, keyword)
		if e != nil {
			return nil, e
		}
		return &listResult{List: list, Total: total}, nil
	})

	if err != nil {
		slog.Error("get resource list failed", "error", err)
		return nil, 0, errors.New("获取资源列表失败")
	}

	if !found || data == nil {
		return []model.Resource{}, 0, nil
	}

	if lr, ok := data.(*listResult); ok {
		return lr.List, lr.Total, nil
	}

	return []model.Resource{}, 0, nil
}

// GetResourceByID 根据 ID 获取资源详情（走热点数据缓存）。
//
// 缓存策略：
//   - 缓存类型：热点数据缓存
//   - 过期方式：逻辑过期 + 异步重建（防击穿）
//   - TTL：30分钟逻辑过期时间，物理永不过期
//   - 缓存穿透：空值标记 + 布隆过滤器
//
// 参数：
//   - id: 资源主键 ID
//
// 返回：资源指针对象，不存在时返回错误。
func (s *ResourceService) GetResourceByID(id uint) (*model.Resource, error) {
	cacheKey := cache.CacheKeyUint("resource", "id", id)
	var result model.Resource

	opt := cache.HotDataOptions()
	opt.BloomModule = "resource"
	data, found, err := s.cache.Fetch(cacheKey, opt, &result, func() (interface{}, error) {
		slog.Debug("cache miss: load resource from db", "resource_id", id)
		resource, e := s.resourceRepo.FindByID(id)
		if e != nil {
			return nil, nil
		}
		return resource, nil
	})

	if err != nil {
		slog.Error("get resource failed", "error", err, "resource_id", id)
		return nil, errors.New("获取资源详情失败")
	}

	if !found || data == nil {
		return nil, errors.New("资源不存在")
	}

	if resource, ok := data.(*model.Resource); ok {
		return resource, nil
	}

	return nil, errors.New("资源不存在")
}

// GetActiveResources 获取所有可用资源列表（用于资源清单展示）。
//
// 缓存策略：
//   - 缓存类型：查询级缓存
//   - TTL：5分钟 ± 10% 抖动
//
// 返回：所有状态为可用的资源列表。
func (s *ResourceService) GetActiveResources() ([]model.Resource, error) {
	cacheKey := cache.CacheKey("resource", "active_list", "all")
	var result []model.Resource

	opt := cache.DefaultFetchOptions(cache.TTLQuery)
	opt.UseBloom = false
	data, found, err := s.cache.Fetch(cacheKey, opt, &result, func() (interface{}, error) {
		slog.Debug("cache miss: load active resources from db")
		return s.resourceRepo.ListActiveResources()
	})

	if err != nil {
		slog.Error("get active resources failed", "error", err)
		return nil, errors.New("获取可用资源列表失败")
	}

	if !found || data == nil {
		return []model.Resource{}, nil
	}

	if resources, ok := data.(*[]model.Resource); ok {
		return *resources, nil
	}

	return []model.Resource{}, nil
}

// CreateResource 创建新资源（管理员操作）。
//
// 业务逻辑：
//  1. 生成 UUID 作为资源唯一标识
//  2. 设置默认状态为可用
//  3. 写入数据库
//  4. 主动失效所有资源列表缓存
//
// 参数：
//   - name: 资源名称
//   - resourceType: 资源类型（llm_chat/llm_code/image_gen/asr/tts/embedding/other）
//   - apiName: API 名称（唯一标识）
//   - description: 资源描述
//   - provider: 提供厂商
//   - version: 版本号
//   - defaultQPS: 默认 QPS 配额
//   - maxQPS: 最大 QPS 配额
//   - docsURL: 文档链接
//   - tags: 标签（JSON 数组字符串）
//
// 返回：创建后的资源对象和错误。
func (s *ResourceService) CreateResource(
	name, resourceType, apiName, description, provider, version string,
	defaultQPS, maxQPS int, docsURL, tags string,
) (*model.Resource, error) {
	now := time.Now()
	resource := &model.Resource{
		UUID:        uuid.New().String(),
		Name:        name,
		Type:        resourceType,
		APIName:     apiName,
		Description: description,
		Provider:    provider,
		Version:     version,
		DefaultQPS:  defaultQPS,
		MaxQPS:      maxQPS,
		Status:      model.ResourceStatusActive,
		DocsURL:     docsURL,
		Tags:        tags,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.resourceRepo.Create(resource); err != nil {
		slog.Error("create resource failed", "error", err, "api_name", apiName)
		return nil, errors.New("创建资源失败")
	}

	slog.Info("resource created", "resource_id", resource.ID, "name", name)
	s.invalidateResourceListCaches()

	return resource, nil
}

// UpdateResource 更新资源信息（管理员操作）。
//
// 业务逻辑：
//  1. 查询原资源（确保存在）
//  2. 更新字段
//  3. 保存到数据库
//  4. 删除详情缓存，失效列表缓存
//
// 参数：
//   - id: 资源主键 ID
//   - name/resourceType/apiName/...: 更新字段
//
// 返回：更新后的资源对象和错误。
func (s *ResourceService) UpdateResource(
	id uint,
	name, resourceType, apiName, description, provider, version string,
	defaultQPS, maxQPS, status int, docsURL, tags string,
) (*model.Resource, error) {
	existing, err := s.resourceRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("资源不存在")
	}

	existing.Name = name
	existing.Type = resourceType
	existing.APIName = apiName
	existing.Description = description
	existing.Provider = provider
	existing.Version = version
	existing.DefaultQPS = defaultQPS
	existing.MaxQPS = maxQPS
	existing.Status = status
	existing.DocsURL = docsURL
	existing.Tags = tags
	existing.UpdatedAt = time.Now()

	if err := s.resourceRepo.Update(existing); err != nil {
		slog.Error("update resource failed", "error", err, "resource_id", id)
		return nil, errors.New("更新资源失败")
	}

	slog.Info("resource updated", "resource_id", id, "name", name)

	// 删除单条详情缓存
	detailKey := cache.CacheKeyUint("resource", "id", id)
	s.cache.Delete(detailKey)
	// 失效列表缓存
	s.invalidateResourceListCaches()

	return existing, nil
}

// DeleteResource 删除资源（管理员操作，软删除）。
//
// 业务逻辑：
//  1. 查询原资源（确保存在）
//  2. 执行软删除（GORM DeletedAt）
//  3. 删除详情缓存，失效列表缓存
//
// 注意：软删除的数据不会出现在正常查询中，但数据库仍保留记录。
//
// 参数：
//   - id: 资源主键 ID
//
// 返回：错误。
func (s *ResourceService) DeleteResource(id uint) error {
	_, err := s.resourceRepo.FindByID(id)
	if err != nil {
		return errors.New("资源不存在")
	}

	if err := s.resourceRepo.Delete(id); err != nil {
		slog.Error("delete resource failed", "error", err, "resource_id", id)
		return errors.New("删除资源失败")
	}

	slog.Info("resource deleted", "resource_id", id)

	// 删除单条详情缓存
	detailKey := cache.CacheKeyUint("resource", "id", id)
	s.cache.Delete(detailKey)
	// 失效列表缓存
	s.invalidateResourceListCaches()

	return nil
}
