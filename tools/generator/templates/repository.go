package templates

const RepositoryTemplate = `package repository

import (
	"context"
	"{{.ModuleName}}/internal/entity"
	"{{.ModuleName}}/pkg/query"

	"{{.ModuleName}}/pkg/redis"
	"gorm.io/gorm"
)

type {{.Name}}Repository interface {
	Create(ctx context.Context, {{.VarName}} *entity.{{.Name}}) error
	FindByID(ctx context.Context, id uint) (*entity.{{.Name}}, error)
	Update(ctx context.Context, {{.VarName}} *entity.{{.Name}}) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *query.PageRequest, queryBuilder query.QueryBuilder) (*query.PageResponse[entity.{{.Name}}], error)
}

type {{.VarName}}Repository struct {
	db *gorm.DB
	cache *redis.Client
}

func New{{.Name}}Repository(db *gorm.DB, cache *redis.Client) {{.Name}}Repository {
	return &{{.VarName}}Repository{db: db, cache: cache}
}

func (r *{{.VarName}}Repository) Create(ctx context.Context, {{.VarName}} *entity.{{.Name}}) error {
	return r.db.WithContext(ctx).Create({{.VarName}}).Error
}

func (r *{{.VarName}}Repository) FindByID(ctx context.Context, id uint) (*entity.{{.Name}}, error) {
	var {{.VarName}} entity.{{.Name}}
	err := r.db.WithContext(ctx).First(&{{.VarName}}, id).Error
	if err != nil {
		return nil, err
	}
	return &{{.VarName}}, nil
}

func (r *{{.VarName}}Repository) Update(ctx context.Context, {{.VarName}} *entity.{{.Name}}) error {
	return r.db.WithContext(ctx).Save({{.VarName}}).Error
}

func (r *{{.VarName}}Repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.{{.Name}}{}, id).Error
}

func (r *{{.VarName}}Repository) List(ctx context.Context, req *query.PageRequest, queryBuilder query.QueryBuilder) (*query.PageResponse[entity.{{.Name}}], error) {
	var {{.VarName}}s []entity.{{.Name}}
	
	// 使用传入的查询构建器
	// 如果查询构建器为空，创建一个新的
	if queryBuilder == nil {
		// 创建一个新的查询构建器
		factory := query.NewMySQLQueryFactory(r.db)
		queryBuilder = factory.NewQuery()
		
		// 设置模型
		db := r.db.WithContext(ctx).Model(&entity.{{.Name}}{})
		queryBuilder.WhereRaw(db)
	}
	
	// 创建数据提供者
	provider := query.NewMySQLProvider[entity.{{.Name}}](r.db)
	
	// 执行分页查询
	return query.Paginate(ctx, queryBuilder, provider, req, &{{.VarName}}s)
}
`
