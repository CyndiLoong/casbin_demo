// role_repo.go 封装 roles 表及 role_permissions 关联表的数据访问操作。
package repository

import (
	"casbin-demo/internal/model"

	"gorm.io/gorm"
)

// RoleRepository 角色数据访问对象。
// 封装 roles 表及 role_permissions 关联表的 CRUD 操作。
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository 创建 RoleRepository 实例。
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create 创建新角色。
func (r *RoleRepository) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

// FindByID 根据主键 ID 查询角色（预加载 Permissions 关联）。
func (r *RoleRepository) FindByID(id uint) (*model.Role, error) {
	var role model.Role
	if err := r.db.Preload("Permissions").First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// FindByName 根据角色名称查询角色（用于唯一性校验和 Casbin 策略加载）。
func (r *RoleRepository) FindByName(name string) (*model.Role, error) {
	var role model.Role
	if err := r.db.Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// List 查询所有角色（预加载 Permissions 关联）。
func (r *RoleRepository) List() ([]model.Role, error) {
	var roles []model.Role
	if err := r.db.Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// Update 更新角色信息（全量保存）。
func (r *RoleRepository) Update(role *model.Role) error {
	return r.db.Save(role).Error
}

// Delete 根据 ID 软删除角色（使用 GORM 软删除机制）。
func (r *RoleRepository) Delete(id uint) error {
	return r.db.Delete(&model.Role{}, id).Error
}

// AssignPermission 为角色分配权限（幂等，ON CONFLICT DO NOTHING）。
func (r *RoleRepository) AssignPermission(roleID, permissionID uint) error {
	return r.db.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING", roleID, permissionID).Error
}

// RemovePermission 移除角色的指定权限。
func (r *RoleRepository) RemovePermission(roleID, permissionID uint) error {
	return r.db.Exec("DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?", roleID, permissionID).Error
}

// Count 查询角色总数（用于仪表盘统计）。
func (r *RoleRepository) Count() (int64, error) {
	var total int64
	if err := r.db.Model(&model.Role{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
