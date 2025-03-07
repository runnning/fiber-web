package query

import "context"

// Querier 通用查询接口
type Querier[T any] interface {
	// FindPage 分页查询
	// 参数：
	//   - ctx: 上下文
	//   - q: 查询参数，包含分页、条件、排序等信息
	// 返回：
	//   - *PageResult[T]: 分页结果，包含数据列表和分页信息
	//   - error: 查询过程中的错误信息
	FindPage(ctx context.Context, q *Query) (*PageResult[T], error)

	// Count 获取总数
	// 参数：
	//   - ctx: 上下文
	//   - q: 查询参数
	// 返回：
	//   - int64: 记录总数
	//   - error: 查询过程中的错误信息
	Count(ctx context.Context, q *Query) (int64, error)

	// First 获取单条记录
	// 参数：
	//   - ctx: 上下文
	//   - q: 查询参数
	// 返回：
	//   - *T: 记录指针
	//   - error: 查询过程中的错误信息
	First(ctx context.Context, q *Query) (*T, error)

	// Find 获取记录列表（不分页）
	// 参数：
	//   - ctx: 上下文
	//   - q: 查询参数
	// 返回：
	//   - []T: 记录列表
	//   - error: 查询过程中的错误信息
	Find(ctx context.Context, q *Query) ([]T, error)
}

// BaseQuerier 基础查询实现
// 现在支持泛型
type BaseQuerier[T any] struct{}

// HandlePagination 处理分页参数
// 返回规范化后的页码和每页数量
func (b *BaseQuerier[T]) HandlePagination(pagination *Pagination) (page, pageSize int) {
	if pagination == nil {
		return DefaultPage, DefaultPageSize
	}
	return NewPagination(pagination.Page, pagination.PageSize).Page,
		NewPagination(pagination.Page, pagination.PageSize).PageSize
}

// CalculateTotalPages 计算总页数
// 参数：
//   - total: 总记录数
//   - pageSize: 每页数量
//
// 返回：总页数
func (b *BaseQuerier[T]) CalculateTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	return totalPages
}

// NewPageResult 创建分页结果
// 参数：
//   - list: 数据列表
//   - total: 总记录数
//   - page: 当前页码
//   - pageSize: 每页数量
//
// 返回：分页结果对象
func (b *BaseQuerier[T]) NewPageResult(list []T, total int64, page, pageSize int) *PageResult[T] {
	// 确保页码和每页数量合法
	if page <= 0 {
		page = DefaultPage
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}

	return &PageResult[T]{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: b.CalculateTotalPages(total, pageSize),
	}
}
