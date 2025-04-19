package entity

const (
	AdminRole    = "admin"
	AdminUserID  = "1"
	MenuResource = "menu:"
	ApiResource  = "api:"
	PermSep      = ","
)

type AdminUserDataItem struct {
	ID        uint     `json:"id"`
	Name      string   `json:"name"`
	Password  string   `json:"password"`
	Email     string   `json:"email"`
	Phone     string   `form:"phone"`
	Roles     []string `json:"roles"`
	UpdatedAt string   `json:"updated_at"`
	CreatedAt string   `json:"created_at"`
}

type GetAdminUserResponseData struct {
	ID        uint     `json:"id"`
	Name      string   `json:"name"`
	Password  string   `json:"password" `
	Email     string   `json:"email" `
	Phone     string   `json:"phone" `
	Roles     []string `json:"roles" `
	UpdatedAt string   `json:"updated_at"`
	CreatedAt string   `json:"created_at"`
}

type GetAdminUserResponse struct {
	Data GetAdminUserResponseData
}

type GetAdminUsersResponseData struct {
	List  []AdminUserDataItem `json:"list"`
	Total int64               `json:"total"`
}

type GetAdminUsersResponse struct {
	Data GetAdminUsersResponseData
}
