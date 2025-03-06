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
	List(ctx context.Context, req *query.PageRequest) (*query.PageResponse[entity.{{.Name}}], error)
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

func (r *{{.VarName}}Repository) List(ctx context.Context, req *query.PageRequest) (*query.PageResponse[entity.{{.Name}}], error) {
	var {{.VarName}}s []entity.{{.Name}}
	db := r.db.WithContext(ctx).Model(&entity.{{.Name}}{})
	
	// 处理搜索条件
	if search := req.GetFilter("search"); search != "" {
		db = query.BuildSearchQuery(db, search, []string{"name"})
	}
	
	// 处理状态过滤
	if status := req.GetFilter("status"); status != "" {
		db = db.Where("status = ?", status)
	}
	
	// 处理时间范围
	startTime := req.GetFilter("start_time")
	endTime := req.GetFilter("end_time")
	if startTime != "" || endTime != "" {
		db = query.BuildTimeRangeQuery(db, "created_at", startTime, endTime)
	}
	
	// 构建查询
	builder := query.NewMySQLQuery(db)
	
	// 添加其他条件
	if category := req.GetFilter("category"); category != "" {
		builder.AddCondition("category", query.OpEq, category)
	}
	
	// 创建数据提供者
	provider := query.NewMySQLProvider[entity.{{.Name}}](r.db)
	
	// 执行分页查询
	return query.Paginate(ctx, builder, provider, req, &{{.VarName}}s)
}
`
