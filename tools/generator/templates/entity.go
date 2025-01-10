package templates

var EntityTemplate = `package entity

import (
	"time"

	"gorm.io/gorm"
)

// {{.Name}} 实体模型
type {{.Name}} struct {
	{{- range .Fields}}
	{{.Name}} {{.Type}} {{.Tag}} // {{.Comment}}
	{{- end}}
}

// TableName 指定表名
func ({{.VarName}} *{{.Name}}) TableName() string {
	return "{{.VarName}}s"
}
`
