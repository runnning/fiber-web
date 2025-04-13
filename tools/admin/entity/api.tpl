package entity

import (
	"time"

	"gorm.io/gorm"
)

// Api 实体模型
type Api struct {
	Id uint  // 主键ID
	ApiGroup string  // api分组
	Name string  // api分组名称
	Path string  // api路径
	Method string  // http方法
	CreatedAt time.Time  // 创建时间
	UpdatedAt time.Time  // 更新时间
	DeletedAt gorm.DeletedAt  // 删除时间
}

// TableName 指定表名
func (api *Api) TableName() string {
	return "apis"
}
