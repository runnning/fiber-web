package templates

var EntityTemplate = `package entity

import (
	"time"

	"gorm.io/gorm"
)

type {{.Name}} struct {
	ID        uint           ` + "`json:\"id\" gorm:\"primarykey\"`" + `
	CreatedAt time.Time      ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time      ` + "`json:\"updated_at\"`" + `
	DeletedAt gorm.DeletedAt ` + "`json:\"deleted_at,omitempty\" gorm:\"index\"`" + `
}
`
