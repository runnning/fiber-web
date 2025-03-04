package query

import (
	"fmt"

	"gorm.io/gorm"
)

// BuildQuery 构建查询接口
type BuildQuery interface {
	Build() interface{}
}

// MySQLQueryBuilder MySQL查询构建器
type MySQLQueryBuilder struct {
	db *gorm.DB
}

// NewMySQLQueryBuilder 创建MySQL查询构建器实例
func NewMySQLQueryBuilder(db *gorm.DB) *MySQLQueryBuilder {
	return &MySQLQueryBuilder{db: db}
}

// Where 添加查询条件
func (b *MySQLQueryBuilder) Where(query interface{}, args ...interface{}) *MySQLQueryBuilder {
	b.db = b.db.Where(query, args...)
	return b
}

// Like 添加模糊查询
func (b *MySQLQueryBuilder) Like(field string, value string) *MySQLQueryBuilder {
	if value != "" {
		b.db = b.db.Where(fmt.Sprintf("%s LIKE ?", field), "%"+value+"%")
	}
	return b
}

// Equal 添加等值查询
func (b *MySQLQueryBuilder) Equal(field string, value interface{}) *MySQLQueryBuilder {
	if value != nil && value != "" {
		b.db = b.db.Where(fmt.Sprintf("%s = ?", field), value)
	}
	return b
}

// In 添加IN查询
func (b *MySQLQueryBuilder) In(field string, values ...interface{}) *MySQLQueryBuilder {
	if len(values) > 0 {
		b.db = b.db.Where(fmt.Sprintf("%s IN ?", field), values)
	}
	return b
}

// Between 添加范围查询
func (b *MySQLQueryBuilder) Between(field string, start, end interface{}) *MySQLQueryBuilder {
	if start != nil {
		b.db = b.db.Where(fmt.Sprintf("%s >= ?", field), start)
	}
	if end != nil {
		b.db = b.db.Where(fmt.Sprintf("%s <= ?", field), end)
	}
	return b
}

// Build 构建查询
func (b *MySQLQueryBuilder) Build() interface{} {
	return b.db
}

// MySQLPaginate MySQL分页查询
func MySQLPaginate[T any](builder *MySQLQueryBuilder, req *PageRequest, result *[]T) (*PageResponse[T], error) {
	var total int64
	db := builder.db

	// 计算总记录数
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	// 排序
	if req.OrderBy != "" {
		order := req.OrderBy
		if req.Order != "" {
			order += " " + req.Order
		}
		db = db.Order(order)
	}

	// 分页查询
	if err := db.Offset(req.Offset()).Limit(req.PageSize).Find(result).Error; err != nil {
		return nil, err
	}

	return NewPageResponse(*result, total, req.Page, req.PageSize), nil
}

// Example usage:
/*
func GetUserList(db *gorm.DB, req *PageRequest) (*PageResponse, error) {
    var users []User

    // 创建查询构建器
    query := NewMySQLQuery(db.Model(&User{})).
        Select("id", "name", "email", "age").
        AddCondition("age", OpGte, 18).
        AddCondition("status", OpEq, "active").
        AddArrayCondition("role", OpIn, []string{"user", "admin"})

    return MySQLPaginate(query, req, &users)
}

// 支持连接查询
func GetUserOrders(db *gorm.DB, req *PageRequest) (*PageResponse, error) {
    var results []struct {
        UserID    uint
        UserName  string
        OrderCount int
        TotalAmount float64
    }

    query := NewMySQLQuery(db.Model(&User{})).
        Select("users.id as user_id", "users.name as user_name",
               "COUNT(orders.id) as order_count",
               "SUM(orders.amount) as total_amount").
        Join("LEFT JOIN orders ON users.id = orders.user_id").
        GroupBy("users.id").
        Having("COUNT(orders.id) > 0")

    return MySQLPaginate(query, req, &results)
}

// 支持时间范围查询
func GetUsersByTimeRange(db *gorm.DB, req *PageRequest, startTime, endTime *time.Time) (*PageResponse, error) {
    var users []User

    query := NewMySQLQuery(db.Model(&User{}))
    if startTime != nil {
        query.AddCondition("created_at", OpGte, startTime)
    }
    if endTime != nil {
        query.AddCondition("created_at", OpLte, endTime)
    }

    return MySQLPaginate(query, req, &users)
}

// 支持多字段模糊搜索
func SearchUsers(db *gorm.DB, req *PageRequest, searchText string) (*PageResponse, error) {
    var users []User

    baseQuery := db.Model(&User{})
    if searchText != "" {
        baseQuery = BuildSearchQuery(baseQuery, searchText, []string{"name", "email", "description"})
    }

    query := NewMySQLQuery(baseQuery)
    return MySQLPaginate(query, req, &users)
}
*/
