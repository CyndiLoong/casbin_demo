// permission_repo.go 封装 permissions 表的数据访问操作。
package repository

import (
	"casbin-demo/internal/model"

	"gorm.io/gorm"
)

// PermissionRepository 权限数据访问对象。
// 封装 permissions 表的 CRUD 操作。
type PermissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository 创建 PermissionRepository 实例。
func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// Create 创建新权限。
func (r *PermissionRepository) Create(permission *model.Permission) error {
	return r.db.Create(permission).Error
}

// FindByID 根据主键 ID 查询权限。
func (r *PermissionRepository) FindByID(id uint) (*model.Permission, error) {
	var permission model.Permission
	if err := r.db.First(&permission, id).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

// List 查询所有权限。
func (r *PermissionRepository) List() ([]model.Permission, error) {
	var permissions []model.Permission
	if err := r.db.Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// Update 更新权限信息（全量保存）。
func (r *PermissionRepository) Update(permission *model.Permission) error {
	return r.db.Save(permission).Error
}

// Delete 根据 ID 软删除权限（使用 GORM 软删除机制）。
func (r *PermissionRepository) Delete(id uint) error {
	return r.db.Delete(&model.Permission{}, id).Error
}

// Count 查询权限总数（用于仪表盘统计）。
func (r *PermissionRepository) Count() (int64, error) {
	var total int64
	if err := r.db.Model(&model.Permission{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
