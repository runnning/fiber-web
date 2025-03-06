package templates

var UseCaseTemplate = `package usecase

import (
	"context"
	"fmt"
	"time"
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
	List(ctx context.Context, req *query.PageRequest, queryBuilder query.QueryBuilder) (*query.PageResponse[entity.{{.Name}}], error)
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

func (uc *{{.VarName}}UseCase) List(ctx context.Context, req *query.PageRequest, queryBuilder query.QueryBuilder) (*query.PageResponse[entity.{{.Name}}], error) {
	// 处理查询参数和业务逻辑
	
	// 参数验证
	//if req.Page <= 0 {
	//	req.Page = 1
	//}
	//if req.PageSize <= 0 || req.PageSize > 100 {
	//	req.PageSize = 10 // 限制最大页面大小
	//}
	
	// 设置默认排序
	if req.OrderBy == "" {
		req.OrderBy = "id"
		req.Order = "DESC"
	}
	
	// 处理业务相关的过滤条件
	if status := req.GetFilter("status"); status != "" {
		// 验证状态值是否有效
		validStatus := map[string]bool{"0": true, "1": true, "2": true}
		if !validStatus[status] {
			return nil, fmt.Errorf("无效的状态值: %s", status)
		}
	}
	
	// 处理时间范围
	startTime := req.GetFilter("start_time")
	endTime := req.GetFilter("end_time")
	if startTime != "" || endTime != "" {
		// 验证时间格式
		if startTime != "" {
			if _, err := time.Parse("2006-01-02", startTime); err != nil {
				return nil, fmt.Errorf("开始时间格式错误: %s", startTime)
			}
		}
		if endTime != "" {
			if _, err := time.Parse("2006-01-02", endTime); err != nil {
				return nil, fmt.Errorf("结束时间格式错误: %s", endTime)
			}
		}
	}
	
	// 对查询构建器进行业务逻辑相关的修改
	if queryBuilder != nil {
		// 例如：根据用户角色添加额外的查询条件
		// 如果当前用户不是管理员，可能需要限制只能查看特定状态的记录
		// queryBuilder.WhereSimple("status", query.OpEq, "active")
		
		// 或者添加默认的排序条件
		// queryBuilder.OrderBy("created_at", "DESC")
		
		// 或者添加安全相关的条件，如排除敏感记录
		// queryBuilder.WhereSimple("is_sensitive", query.OpEq, false)
	}
	
	// 调用仓库层执行查询
	result, err := uc.{{.VarName}}Repo.List(ctx, req, queryBuilder)
	if err != nil {
		return nil, err
	}
	
	// 对结果进行后处理
	// 例如：敏感信息过滤、数据转换等
	
	return result, nil
}
`
