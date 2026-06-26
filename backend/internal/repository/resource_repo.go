// resource_repo.go 封装 api_resources 表的数据访问操作。
//
// 资源清单用于展示所有可申请的大模型API资源，
// 用户可查看资源详情并提交申请。
package repository

import (
	"casbin-demo/internal/model"

	"gorm.io/gorm"
)

// ResourceRepository 资源数据访问对象。
// 封装 api_resources 表的 CRUD 操作。
type ResourceRepository struct {
	db *gorm.DB
}

// NewResourceRepository 创建 ResourceRepository 实例。
func NewResourceRepository(db *gorm.DB) *ResourceRepository {
	return &ResourceRepository{db: db}
}

// AutoMigrate 自动迁移资源表结构。
// 使用 GORM AutoMigrate 自动创建/更新表结构和索引。
func (r *ResourceRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&model.Resource{})
}

// Create 创建新资源。
func (r *ResourceRepository) Create(resource *model.Resource) error {
	return r.db.Create(resource).Error
}

// FindByID 根据主键 ID 查询资源。
func (r *ResourceRepository) FindByID(id uint) (*model.Resource, error) {
	var resource model.Resource
	if err := r.db.First(&resource, id).Error; err != nil {
		return nil, err
	}
	return &resource, nil
}

// FindByUUID 根据 UUID 查询资源。
func (r *ResourceRepository) FindByUUID(uuid string) (*model.Resource, error) {
	var resource model.Resource
	if err := r.db.Where("uuid = ?", uuid).First(&resource).Error; err != nil {
		return nil, err
	}
	return &resource, nil
}

// List 分页查询资源列表。
//
// 参数：
//   - page: 页码（从 1 开始）
//   - pageSize: 每页数量
//   - resourceType: 资源类型筛选（空字符串表示不筛选）
//   - status: 状态筛选（nil 表示不筛选）
//   - keyword: 关键词搜索（匹配名称/API名称/描述）
//
// 返回：资源列表、总记录数、错误。
func (r *ResourceRepository) List(page, pageSize int, resourceType string, status *int, keyword string) ([]model.Resource, int64, error) {
	var resources []model.Resource
	var total int64

	query := r.db.Model(&model.Resource{})

	if resourceType != "" {
		query = query.Where("type = ?", resourceType)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if keyword != "" {
		keywordPattern := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR api_name LIKE ? OR description LIKE ?",
			keywordPattern, keywordPattern, keywordPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&resources).Error; err != nil {
		return nil, 0, err
	}

	return resources, total, nil
}

// Update 更新资源信息（全量保存）。
func (r *ResourceRepository) Update(resource *model.Resource) error {
	return r.db.Save(resource).Error
}

// Delete 根据 ID 软删除资源。
func (r *ResourceRepository) Delete(id uint) error {
	return r.db.Delete(&model.Resource{}, id).Error
}

// Count 查询资源总数。
func (r *ResourceRepository) Count() (int64, error) {
	var total int64
	if err := r.db.Model(&model.Resource{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// ListActiveResources 查询所有可用资源（用于资源清单展示）。
//
// 只返回状态为可用的资源，按创建时间倒序排列。
func (r *ResourceRepository) ListActiveResources() ([]model.Resource, error) {
	var resources []model.Resource
	if err := r.db.Where("status = ?", model.ResourceStatusActive).
		Order("created_at DESC").
		Find(&resources).Error; err != nil {
		return nil, err
	}
	return resources, nil
}
