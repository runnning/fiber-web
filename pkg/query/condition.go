package query

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// Option 查询选项接口
type Option interface {
	Apply(*gorm.DB) *gorm.DB
}

// Condition 查询条件
type Condition struct {
	Field    string
	Operator string
	Value    interface{}
}

// Apply 实现 Option 接口
func (c Condition) Apply(db *gorm.DB) *gorm.DB {
	switch strings.ToUpper(c.Operator) {
	case "=":
		return db.Where(fmt.Sprintf("%s = ?", c.Field), c.Value)
	case "!=", "<>":
		return db.Where(fmt.Sprintf("%s != ?", c.Field), c.Value)
	case ">":
		return db.Where(fmt.Sprintf("%s > ?", c.Field), c.Value)
	case ">=":
		return db.Where(fmt.Sprintf("%s >= ?", c.Field), c.Value)
	case "<":
		return db.Where(fmt.Sprintf("%s < ?", c.Field), c.Value)
	case "<=":
		return db.Where(fmt.Sprintf("%s <= ?", c.Field), c.Value)
	case "LIKE":
		return db.Where(fmt.Sprintf("%s LIKE ?", c.Field), c.Value)
	case "IN":
		return db.Where(fmt.Sprintf("%s IN (?)", c.Field), c.Value)
	case "NOT IN":
		return db.Where(fmt.Sprintf("%s NOT IN (?)", c.Field), c.Value)
	case "BETWEEN":
		values, ok := c.Value.([]interface{})
		if ok && len(values) == 2 {
			return db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", c.Field), values[0], values[1])
		}
	case "NULL":
		return db.Where(fmt.Sprintf("%s IS NULL", c.Field))
	case "NOT NULL":
		return db.Where(fmt.Sprintf("%s IS NOT NULL", c.Field))
	}
	return db
}

// Order 排序条件
type Order struct {
	Field string
	Desc  bool
}

// Apply 实现 Option 接口
func (o Order) Apply(db *gorm.DB) *gorm.DB {
	direction := "ASC"
	if o.Desc {
		direction = "DESC"
	}
	return db.Order(fmt.Sprintf("%s %s", o.Field, direction))
}

// Select 字段选择
type Select struct {
	Fields []string
}

// Apply 实现 Option 接口
func (s Select) Apply(db *gorm.DB) *gorm.DB {
	return db.Select(s.Fields)
}

// WithOptions 应用查询选项
func WithOptions(db *gorm.DB, opts ...Option) *gorm.DB {
	for _, opt := range opts {
		db = opt.Apply(db)
	}
	return db
}
