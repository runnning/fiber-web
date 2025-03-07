package query

const (
	// DefaultPage 默认页码
	DefaultPage = 1
	// DefaultPageSize 默认每页数量
	DefaultPageSize = 10
)

// Pagination 分页参数
type Pagination struct {
	Page     int `json:"page" form:"page"`         // 当前页码
	PageSize int `json:"pageSize" form:"pageSize"` // 每页数量
}

// NewPagination 创建分页参数
func NewPagination(page, pageSize int) *Pagination {
	if page <= 0 {
		page = DefaultPage
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	return &Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

// Operator 查询操作符
type Operator string

const (
	OpEq         Operator = "eq"          // 等于
	OpNe         Operator = "ne"          // 不等于
	OpGt         Operator = "gt"          // 大于
	OpGte        Operator = "gte"         // 大于等于
	OpLt         Operator = "lt"          // 小于
	OpLte        Operator = "lte"         // 小于等于
	OpIn         Operator = "in"          // 在列表中
	OpNotIn      Operator = "not_in"      // 不在列表中
	OpLike       Operator = "like"        // 模糊匹配
	OpNotLike    Operator = "not_like"    // 不匹配
	OpBetween    Operator = "between"     // 区间
	OpNotBetween Operator = "not_between" // 不在区间
)

// Condition 查询条件
type Condition struct {
	Field    string   `json:"field"`    // 字段名
	Operator Operator `json:"operator"` // 操作符
	Value    any      `json:"value"`    // 值
}

// Query 查询参数
type Query struct {
	Pagination       *Pagination `json:"pagination"`             // 分页参数
	Conditions       []Condition `json:"conditions"`             // 查询条件
	OrderBy          []string    `json:"orderBy,omitempty"`      // 排序字段，格式：字段名 ASC/DESC
	SelectFields     []string    `json:"selectFields,omitempty"` // 选择的字段
	EnablePagination bool        `json:"enablePagination"`       // 是否启用分页
}

// NewQuery 创建新的查询参数
func NewQuery() *Query {
	return &Query{
		Conditions:       make([]Condition, 0, 4),                     // 预分配4个条件的空间
		OrderBy:          make([]string, 0, 2),                        // 预分配2个排序的空间
		SelectFields:     make([]string, 0, 4),                        // 预分配4个字段的空间
		EnablePagination: true,                                        // 默认启用分页
		Pagination:       NewPagination(DefaultPage, DefaultPageSize), // 默认分页参数
	}
}

// SetPage 设置分页参数
func (q *Query) SetPage(page, pageSize int) *Query {
	q.Pagination = NewPagination(page, pageSize)
	return q
}

// SetPagination 设置分页参数
func (q *Query) SetPagination(pagination *Pagination) *Query {
	q.Pagination = pagination
	return q
}

// DisablePagination 禁用分页
func (q *Query) DisablePagination() *Query {
	q.EnablePagination = false
	return q
}

// EnablePaginationFunc 启用分页
func (q *Query) EnablePaginationFunc() *Query {
	q.EnablePagination = true
	return q
}

// Select 设置要查询的字段
func (q *Query) Select(fields ...string) *Query {
	q.SelectFields = fields
	return q
}

// AddSelect 添加要查询的字段
func (q *Query) AddSelect(fields ...string) *Query {
	q.SelectFields = append(q.SelectFields, fields...)
	return q
}

// AddCondition 追加查询条件
func (q *Query) AddCondition(field string, operator Operator, value any) *Query {
	q.Conditions = append(q.Conditions, Condition{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return q
}

// AddOrderBy 追加排序字段
func (q *Query) AddOrderBy(order string) *Query {
	q.OrderBy = append(q.OrderBy, order)
	return q
}

// PageResult 分页结果
type PageResult[T any] struct {
	List       []T   `json:"list"`       // 数据列表
	Total      int64 `json:"total"`      // 总数
	Page       int   `json:"page"`       // 当前页
	PageSize   int   `json:"pageSize"`   // 每页数量
	TotalPages int   `json:"totalPages"` // 总页数
}
