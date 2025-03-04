package query

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// MySQLQuery MySQL查询结构
type MySQLQuery struct {
	db         *gorm.DB          // 原始数据库连接
	conditions []FilterCondition // 查询条件
	fields     []string          // 查询字段
	joins      []string          // 连接查询
	groupBy    string            // 分组
	having     string            // 分组条件
}

// NewMySQLQuery 创建新的MySQL查询
func NewMySQLQuery(db *gorm.DB) *MySQLQuery {
	return &MySQLQuery{
		db:         db,
		conditions: make([]FilterCondition, 0),
		fields:     make([]string, 0),
		joins:      make([]string, 0),
	}
}

// Select 设置查询字段
func (q *MySQLQuery) Select(fields ...string) *MySQLQuery {
	q.fields = append(q.fields, fields...)
	return q
}

// Join 添加连接查询
func (q *MySQLQuery) Join(join string) *MySQLQuery {
	q.joins = append(q.joins, join)
	return q
}

// GroupBy 设置分组
func (q *MySQLQuery) GroupBy(groupBy string) *MySQLQuery {
	q.groupBy = groupBy
	return q
}

// Having 设置分组条件
func (q *MySQLQuery) Having(having string) *MySQLQuery {
	q.having = having
	return q
}

// AddCondition 添加查询条件
func (q *MySQLQuery) AddCondition(field string, op Operator, value interface{}) *MySQLQuery {
	q.conditions = append(q.conditions, FilterCondition{
		Field:    field,
		Operator: op,
		Value:    value,
	})
	return q
}

// AddArrayCondition 添加数组类型的查询条件
func (q *MySQLQuery) AddArrayCondition(field string, op Operator, values []string) *MySQLQuery {
	q.conditions = append(q.conditions, FilterCondition{
		Field:    field,
		Operator: op,
		Values:   values,
	})
	return q
}

// buildQuery 构建查询
func (q *MySQLQuery) buildQuery() *gorm.DB {
	query := q.db

	// 添加字段选择
	if len(q.fields) > 0 {
		query = query.Select(q.fields)
	}

	// 添加连接
	for _, join := range q.joins {
		query = query.Joins(join)
	}

	// 添加条件
	for _, condition := range q.conditions {
		switch condition.Operator {
		case OpEq:
			query = query.Where(fmt.Sprintf("%s = ?", condition.Field), condition.Value)
		case OpNe:
			query = query.Where(fmt.Sprintf("%s != ?", condition.Field), condition.Value)
		case OpGt:
			query = query.Where(fmt.Sprintf("%s > ?", condition.Field), condition.Value)
		case OpGte:
			query = query.Where(fmt.Sprintf("%s >= ?", condition.Field), condition.Value)
		case OpLt:
			query = query.Where(fmt.Sprintf("%s < ?", condition.Field), condition.Value)
		case OpLte:
			query = query.Where(fmt.Sprintf("%s <= ?", condition.Field), condition.Value)
		case OpIn:
			query = query.Where(fmt.Sprintf("%s IN (?)", condition.Field), condition.Values)
		case OpNin:
			query = query.Where(fmt.Sprintf("%s NOT IN (?)", condition.Field), condition.Values)
		case OpContains:
			if strVal, ok := condition.Value.(string); ok {
				query = query.Where(fmt.Sprintf("%s LIKE ?", condition.Field), "%"+strVal+"%")
			}
		case OpStartsWith:
			if strVal, ok := condition.Value.(string); ok {
				query = query.Where(fmt.Sprintf("%s LIKE ?", condition.Field), strVal+"%")
			}
		case OpEndsWith:
			if strVal, ok := condition.Value.(string); ok {
				query = query.Where(fmt.Sprintf("%s LIKE ?", condition.Field), "%"+strVal)
			}
		case OpExists:
			if boolVal, ok := condition.Value.(bool); ok {
				if boolVal {
					query = query.Where(fmt.Sprintf("%s IS NOT NULL", condition.Field))
				} else {
					query = query.Where(fmt.Sprintf("%s IS NULL", condition.Field))
				}
			}
		}
	}

	// 添加分组
	if q.groupBy != "" {
		query = query.Group(q.groupBy)
		if q.having != "" {
			query = query.Having(q.having)
		}
	}

	return query
}

// BuildSearchQuery 构建搜索查询
func BuildSearchQuery(db *gorm.DB, searchText string, fields []string) *gorm.DB {
	if searchText == "" || len(fields) == 0 {
		return db
	}

	var conditions []string
	var values []interface{}
	searchText = strings.TrimSpace(searchText)

	for _, field := range fields {
		conditions = append(conditions, fmt.Sprintf("%s LIKE ?", field))
		values = append(values, "%"+searchText+"%")
	}

	return db.Where(strings.Join(conditions, " OR "), values...)
}

// BuildTimeRangeQuery 构建时间范围查询
func BuildTimeRangeQuery(db *gorm.DB, field string, startTime, endTime interface{}) *gorm.DB {
	if startTime != nil {
		db = db.Where(fmt.Sprintf("%s >= ?", field), startTime)
	}
	if endTime != nil {
		db = db.Where(fmt.Sprintf("%s <= ?", field), endTime)
	}
	return db
}
