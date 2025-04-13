package entity

import (
	"time"

	"gorm.io/gorm"
)

// AdminUser 实体模型
type AdminUser struct {
	Id uint  // 主键ID
	Name string  // 用户名
	CreatedAt time.Time  // 创建时间
	UpdatedAt time.Time  // 更新时间
	DeletedAt gorm.DeletedAt  // 删除时间
}

// TableName 指定表名
func (admin_user *AdminUser) TableName() string {
	return "admin_users"
}
