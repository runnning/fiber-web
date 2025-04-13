package entity

import (
	"time"

	"gorm.io/gorm"
)

// Role 实体模型
type Role struct {
	Id uint  // 主键ID
	Name string  // 角色名
	CreatedAt time.Time  // 创建时间
	UpdatedAt time.Time  // 更新时间
	DeletedAt gorm.DeletedAt  // 删除时间
}

// TableName 指定表名
func (role *Role) TableName() string {
	return "roles"
}
