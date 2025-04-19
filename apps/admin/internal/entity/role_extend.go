package entity

type RoleDataItem struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Sid       string `json:"sid"`
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}

type GetRolesResponseData struct {
	List  []RoleDataItem `json:"list"`
	Total int64          `json:"total"`
}
type GetRolesResponse struct {
	Data GetRolesResponseData
}
