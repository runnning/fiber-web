package entity

import (
	"time"

	"gorm.io/gorm"
)

// User 实体模型
type User struct {
	Id uint  // 主键ID
	Name string  // 用户名
	Email string  // 邮箱
	Password string  // 密码
	Status int8  // 状态
	CreatedAt time.Time  // 创建时间
	UpdatedAt time.Time  // 更新时间
	DeletedAt gorm.DeletedAt  // 删除时间
}

// TableName 指定表名
func (user *User) TableName() string {
	return "users"
}
