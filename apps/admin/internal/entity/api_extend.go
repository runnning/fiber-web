package entity

type ApiDataItem struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Method    string `json:"method"`
	Group     string `json:"group"`
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}
type GetApisResponseData struct {
	List   []ApiDataItem `json:"list"`
	Total  int64         `json:"total"`
	Groups []string      `json:"groups"`
}
type GetApisResponse struct {
	Data GetApisResponseData
}

type GetUserPermissionsData struct {
	List []string `json:"list"`
}

type GetRolePermissionsData struct {
	List []string `json:"list"`
}
