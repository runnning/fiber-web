package query

import (
	"context"
)

// ===== 分页相关 =====

// PageRequest 分页请求
type PageRequest struct {
	Page     int               `json:"page" query:"page"`         // 页码
	PageSize int               `json:"pageSize" query:"pageSize"` // 每页大小
	OrderBy  string            `json:"orderBy" query:"orderBy"`   // 排序字段
	Order    string            `json:"order" query:"order"`       // 排序方向：ASC/DESC
	Filters  map[string]string `json:"filters"`                   // 过滤条件
}

// PageResponse 分页响应
type PageResponse[T any] struct {
	List     []T   `json:"list"`     // 数据列表
	Total    int64 `json:"total"`    // 总记录数
	Page     int   `json:"page"`     // 当前页码
	PageSize int   `json:"pageSize"` // 每页大小
}

// NewPageRequest 创建分页请求
func NewPageRequest(page, pageSize int) *PageRequest {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return &PageRequest{
		Page:     page,
		PageSize: pageSize,
		Filters:  make(map[string]string),
	}
}

// Offset 获取偏移量
func (r *PageRequest) Offset() int {
	return (r.Page - 1) * r.PageSize
}

// AddFilter 添加过滤条件
func (r *PageRequest) AddFilter(key, value string) *PageRequest {
	r.Filters[key] = value
	return r
}

// GetFilter 获取过滤条件
func (r *PageRequest) GetFilter(key string) string {
	return r.Filters[key]
}

// NewPageResponse 创建分页响应
func NewPageResponse[T any](list []T, total int64, page, pageSize int) *PageResponse[T] {
	return &PageResponse[T]{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// ===== 查询接口 =====

// QueryBuilder 通用查询构建器接口
type QueryBuilder interface {
	// Build 构建查询条件
	Build() interface{}
}

// DataProvider 数据提供者接口
type DataProvider[T any] interface {
	// Count 计算符合条件的记录总数
	Count(ctx context.Context, query interface{}) (int64, error)

	// Find 查询数据列表
	Find(ctx context.Context, query interface{}, req *PageRequest, result *[]T) error
}

// Paginate 通用分页查询函数
func Paginate[T any](ctx context.Context, builder QueryBuilder, provider DataProvider[T], req *PageRequest, result *[]T) (*PageResponse[T], error) {
	query := builder.Build()

	// 计算总记录数
	total, err := provider.Count(ctx, query)
	if err != nil {
		return nil, err
	}

	// 查询数据列表
	if err := provider.Find(ctx, query, req, result); err != nil {
		return nil, err
	}

	return NewPageResponse(*result, total, req.Page, req.PageSize), nil
}
