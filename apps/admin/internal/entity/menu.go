package entity

import (
	"time"

	"gorm.io/gorm"
)

// Menu 实体模型
type Menu struct {
	Id        uint           // 主键ID
	ParentId  uint           // 父级ID
	Path      string         // 地址
	Title     string         // 标题
	Name      string         // 路由中的name
	Component string         // 绑定的组件，默认类型：Iframe、RouteView、ComponentError
	Locale    string         // 本地化标识
	Icon      string         // 图标
	Redirect  string         // 重定向地址
	Url       string         // iframe模式下的跳转url，不能与path重复
	KeepAlive int8           // 是否缓存
	HideMenu  int8           // 是否隐藏
	Target    string         // 全连接跳转模式
	Weight    int            // 排序权重
	CreatedAt time.Time      // 创建时间
	UpdatedAt time.Time      // 更新时间
	DeletedAt gorm.DeletedAt // 删除时间
}

// TableName 指定表名
func (menu *Menu) TableName() string {
	return "menus"
}
