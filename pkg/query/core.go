package query

import (
	"context"
	"errors"
	"fmt"
)

// ===== 分页相关 =====

// PageRequest 分页请求
type PageRequest struct {
	Page     int    `json:"page" query:"page"`         // 页码
	PageSize int    `json:"pageSize" query:"pageSize"` // 每页大小
	OrderBy  string `json:"orderBy" query:"orderBy"`   // 排序字段
	Order    string `json:"order" query:"order"`       // 排序方向：ASC/DESC
}

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// Validate 验证并规范化分页请求参数
func (r *PageRequest) Validate() {
	if r.Page <= 0 {
		r.Page = DefaultPage
	}
	if r.PageSize <= 0 {
		r.PageSize = DefaultPageSize
	} else if r.PageSize > MaxPageSize {
		r.PageSize = MaxPageSize
	}
	if r.Order != "" && r.Order != "ASC" && r.Order != "DESC" {
		r.Order = "ASC"
	}
}

// NewPageRequest 创建分页请求
func NewPageRequest(page, pageSize int) *PageRequest {
	req := &PageRequest{
		Page:     page,
		PageSize: pageSize,
	}
	req.Validate()
	return req
}

// PageResponse 分页响应
type PageResponse[T any] struct {
	List     []T   `json:"list"`     // 数据列表
	Total    int64 `json:"total"`    // 总记录数
	Page     int   `json:"page"`     // 当前页码
	PageSize int   `json:"pageSize"` // 每页大小
}

// Offset 获取偏移量
func (r *PageRequest) Offset() int {
	return (r.Page - 1) * r.PageSize
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

// Condition 表示查询条件
type Condition interface {
	// GetType 获取条件类型
	GetType() ConditionType
}

// ConditionType 条件类型
type ConditionType string

const (
	ConditionTypeSimple ConditionType = "simple" // 简单条件
	ConditionTypeGroup  ConditionType = "group"  // 条件组
	ConditionTypeRaw    ConditionType = "raw"    // 原始条件
)

// SimpleCondition 简单条件
type SimpleCondition struct {
	Field    string      // 字段名
	Operator Operator    // 操作符
	Value    interface{} // 值
}

// GetType 获取条件类型
func (c *SimpleCondition) GetType() ConditionType {
	return ConditionTypeSimple
}

// GroupCondition 条件组
type GroupCondition struct {
	Logic      LogicOperator // 逻辑操作符
	Conditions []Condition   // 子条件列表
}

// GetType 获取条件类型
func (c *GroupCondition) GetType() ConditionType {
	return ConditionTypeGroup
}

// RawCondition 原始条件（用于特定数据库的原生查询）
type RawCondition struct {
	Raw interface{} // 原始条件
}

// GetType 获取条件类型
func (c *RawCondition) GetType() ConditionType {
	return ConditionTypeRaw
}

// LogicOperator 逻辑操作符
type LogicOperator string

const (
	LogicAnd LogicOperator = "AND" // 与
	LogicOr  LogicOperator = "OR"  // 或
)

// QueryBuilder 通用查询构建器接口
type QueryBuilder interface {
	// Where 添加条件
	Where(condition Condition) QueryBuilder

	// WhereSimple 添加简单条件
	WhereSimple(field string, op Operator, value interface{}) QueryBuilder

	// WhereIn 添加IN条件
	WhereIn(field string, values []interface{}) QueryBuilder

	// WhereGroup 添加条件组
	WhereGroup(logic LogicOperator, conditions ...Condition) QueryBuilder

	// WhereRaw 添加原始条件
	WhereRaw(raw interface{}) QueryBuilder

	// Select 设置查询字段
	Select(fields ...string) QueryBuilder

	// OrderBy 设置排序
	OrderBy(field string, direction string) QueryBuilder

	// GroupBy 设置分组
	GroupBy(field string) QueryBuilder

	// Having 设置分组条件
	Having(condition Condition) QueryBuilder

	// Limit 设置限制
	Limit(limit int) QueryBuilder

	// Offset 设置偏移
	Offset(offset int) QueryBuilder

	// Join 添加连接
	Join(table string, condition string) QueryBuilder

	// Build 构建查询
	Build() interface{}
}

// DataProvider 数据提供者接口
type DataProvider[T any] interface {
	// Count 计算符合条件的记录总数
	Count(ctx context.Context, query interface{}) (int64, error)

	// Find 查询数据列表
	Find(ctx context.Context, query interface{}, req *PageRequest, result *[]T) error

	// FindOne 查询单条记录
	FindOne(ctx context.Context, query interface{}, result *T) error

	// Insert 插入记录
	Insert(ctx context.Context, data *T) error

	// Update 更新记录
	Update(ctx context.Context, query interface{}, data map[string]interface{}) error

	// Delete 删除记录
	Delete(ctx context.Context, query interface{}) error

	// Transaction 事务操作
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// QueryFactory 查询工厂接口
type QueryFactory interface {
	// NewQuery 创建新的查询构建器
	NewQuery() QueryBuilder
}

// Paginate 通用分页查询函数
func Paginate[T any](ctx context.Context, builder QueryBuilder, provider DataProvider[T], req *PageRequest, result *[]T) (*PageResponse[T], error) {
	if builder == nil || provider == nil || req == nil || result == nil {
		return nil, errors.New("invalid parameters")
	}

	// 验证并规范化分页参数
	req.Validate()

	// 计算总记录数（不应用分页参数）
	countQuery := builder.Build()
	total, err := provider.Count(ctx, countQuery)
	if err != nil {
		return nil, fmt.Errorf("count error: %w", err)
	}

	// 如果没有数据，直接返回空结果
	if total == 0 {
		*result = make([]T, 0)
		return NewPageResponse(*result, total, req.Page, req.PageSize), nil
	}

	// 应用排序
	if req.OrderBy != "" {
		builder.OrderBy(req.OrderBy, req.Order)
	}

	// 设置分页限制和偏移
	builder.Limit(req.PageSize).Offset(req.Offset())

	// 查询数据列表
	query := builder.Build()
	if err := provider.Find(ctx, query, req, result); err != nil {
		return nil, fmt.Errorf("find error: %w", err)
	}

	return NewPageResponse(*result, total, req.Page, req.PageSize), nil
}
