package templates

var UseCaseTemplate = `package usecase

import (
	"context"
	"{{.ModuleName}}/internal/entity"
	"{{.ModuleName}}/internal/repository"
	"{{.ModuleName}}/pkg/query"
)

// {{.Name}}UseCase 用例接口
type {{.Name}}UseCase interface {
	Create{{.Name}}(ctx context.Context, {{.VarName}} *entity.{{.Name}}) error
	Get{{.Name}}(ctx context.Context, id uint) (*entity.{{.Name}}, error)
	Update{{.Name}}(ctx context.Context, {{.VarName}} *entity.{{.Name}}) error
	Delete{{.Name}}(ctx context.Context, id uint) error
	List(ctx context.Context, req *query.PageRequest) (*query.PageResponse[entity.{{.Name}}], error)
}

// {{.VarName}}UseCase 用例实现
type {{.VarName}}UseCase struct {
	{{.VarName}}Repo repository.{{.Name}}Repository
}

// New{{.Name}}UseCase 创建用例实例
func New{{.Name}}UseCase({{.VarName}}Repo repository.{{.Name}}Repository) {{.Name}}UseCase {
	return &{{.VarName}}UseCase{
		{{.VarName}}Repo: {{.VarName}}Repo,
	}
}

func (uc *{{.VarName}}UseCase) Create{{.Name}}(ctx context.Context, {{.VarName}} *entity.{{.Name}}) error {
	return uc.{{.VarName}}Repo.Create(ctx, {{.VarName}})
}

func (uc *{{.VarName}}UseCase) Get{{.Name}}(ctx context.Context, id uint) (*entity.{{.Name}}, error) {
	return uc.{{.VarName}}Repo.FindByID(ctx, id)
}

func (uc *{{.VarName}}UseCase) Update{{.Name}}(ctx context.Context, {{.VarName}} *entity.{{.Name}}) error {
	return uc.{{.VarName}}Repo.Update(ctx, {{.VarName}})
}

func (uc *{{.VarName}}UseCase) Delete{{.Name}}(ctx context.Context, id uint) error {
	return uc.{{.VarName}}Repo.Delete(ctx, id)
}

func (uc *{{.VarName}}UseCase) List(ctx context.Context, req *query.PageRequest) (*query.PageResponse[entity.{{.Name}}], error) {
	return uc.{{.VarName}}Repo.List(ctx, req)
}
`
