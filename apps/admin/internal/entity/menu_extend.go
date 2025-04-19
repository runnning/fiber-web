package entity

type MenuDataItem struct {
	ID         uint   `json:"id"`                     // 唯一id，使用整数表示
	ParentID   uint   `json:"parent_id,omitempty"`    // 父级菜单的id，使用整数表示
	Weight     int    `json:"weight"`                 // 排序权重
	Path       string `json:"path"`                   // 地址
	Title      string `json:"title"`                  // 展示名称
	Name       string `json:"name,omitempty"`         // 同路由中的name，唯一标识
	Component  string `json:"component,omitempty"`    // 绑定的组件
	Locale     string `json:"locale,omitempty"`       // 本地化标识
	Icon       string `json:"icon,omitempty"`         // 图标，使用字符串表示
	Redirect   string `json:"redirect,omitempty"`     // 重定向地址
	KeepAlive  bool   `json:"keep_alive,omitempty"`   // 是否保活
	HideInMenu bool   `json:"hide_in_menu,omitempty"` // 是否保活
	URL        string `json:"url,omitempty"`          // iframe模式下的跳转url，不能与path重复
	UpdatedAt  string `json:"updated_at,omitempty"`   // 是否保活
}

type GetMenuResponseData struct {
	List []MenuDataItem `json:"list"`
}

type GetMenuResponse struct {
	Data GetMenuResponseData
}
