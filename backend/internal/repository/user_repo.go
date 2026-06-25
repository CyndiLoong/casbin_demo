// user_repo.go 封装 users 表及 user_roles 关联表的数据访问操作。
package repository

import (
	"casbin-demo/internal/model"

	"gorm.io/gorm"
)

// UserRepository 用户数据访问对象。
// 封装 users 表及 user_roles 关联表的 CRUD 操作。
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建 UserRepository 实例。
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建新用户。
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// FindByID 根据主键 ID 查询用户（预加载 Roles 关联）。
func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.Preload("Roles").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查询用户（预加载 Roles 关联）。
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Preload("Roles").Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUUID 根据 UUID 查询用户（预加载 Roles 关联）。
func (r *UserRepository) FindByUUID(uuid string) (*model.User, error) {
	var user model.User
	if err := r.db.Preload("Roles").Where("uuid = ?", uuid).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// List 分页查询用户列表。
//
// 参数：
//   - page: 页码（从 1 开始）
//   - pageSize: 每页数量
//
// 返回：用户列表、总记录数、错误。
func (r *UserRepository) List(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	if err := r.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.Preload("Roles").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Update 更新用户信息（全量保存）。
func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// Delete 根据 ID 软删除用户（使用 GORM 软删除机制）。
func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

// AssignRole 为用户分配角色（幂等，ON CONFLICT DO NOTHING）。
// 使用原生 SQL 避免重复插入错误。
func (r *UserRepository) AssignRole(userID, roleID uint) error {
	return r.db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?) ON CONFLICT DO NOTHING", userID, roleID).Error
}

// RemoveRole 移除用户的指定角色。
func (r *UserRepository) RemoveRole(userID, roleID uint) error {
	return r.db.Exec("DELETE FROM user_roles WHERE user_id = ? AND role_id = ?", userID, roleID).Error
}

// Count 查询用户总数（用于仪表盘统计）。
func (r *UserRepository) Count() (int64, error) {
	var total int64
	if err := r.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// FindAdminUserIDs 查询所有管理员用户的 ID 列表。
func (r *UserRepository) FindAdminUserIDs() ([]uint, error) {
	var ids []uint
	if err := r.db.Model(&model.User{}).
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("roles.name = ?", "admin").
		Pluck("users.id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}
