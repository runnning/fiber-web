package query

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

// 操作符常量
const (
	OpEQ      = "="
	OpNE      = "!="
	OpGT      = ">"
	OpGTE     = ">="
	OpLT      = "<"
	OpLTE     = "<="
	OpLIKE    = "LIKE"
	OpIN      = "IN"
	OpNOTIN   = "NOT IN"
	OpBETWEEN = "BETWEEN"
	OpNULL    = "NULL"
	OpNOTNULL = "NOT NULL"
)

// QueryBuilder 查询构建器接口
type QueryBuilder interface {
	MySQLBuilder
	MongoBuilder
}

// MySQLBuilder MySQL查询构建器接口
type MySQLBuilder interface {
	Apply(*gorm.DB) *gorm.DB
}

// MongoBuilder MongoDB查询构建器接口
type MongoBuilder interface {
	ApplyMongo(*options.FindOptions, *bson.D)
}

// Condition 查询条件
type Condition struct {
	Field    string      // 字段名
	Operator string      // 操作符
	Value    interface{} // 值
}

// Apply 应用MySQL查询条件
func (c Condition) Apply(db *gorm.DB) *gorm.DB {
	switch strings.ToUpper(c.Operator) {
	case OpEQ:
		return db.Where(fmt.Sprintf("%s = ?", c.Field), c.Value)
	case OpNE:
		return db.Where(fmt.Sprintf("%s != ?", c.Field), c.Value)
	case OpGT:
		return db.Where(fmt.Sprintf("%s > ?", c.Field), c.Value)
	case OpGTE:
		return db.Where(fmt.Sprintf("%s >= ?", c.Field), c.Value)
	case OpLT:
		return db.Where(fmt.Sprintf("%s < ?", c.Field), c.Value)
	case OpLTE:
		return db.Where(fmt.Sprintf("%s <= ?", c.Field), c.Value)
	case OpLIKE:
		return db.Where(fmt.Sprintf("%s LIKE ?", c.Field), c.Value)
	case OpIN:
		return db.Where(fmt.Sprintf("%s IN (?)", c.Field), c.Value)
	case OpNOTIN:
		return db.Where(fmt.Sprintf("%s NOT IN (?)", c.Field), c.Value)
	case OpBETWEEN:
		if values, ok := c.Value.([]interface{}); ok && len(values) == 2 {
			return db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", c.Field), values[0], values[1])
		}
	case OpNULL:
		return db.Where(fmt.Sprintf("%s IS NULL", c.Field))
	case OpNOTNULL:
		return db.Where(fmt.Sprintf("%s IS NOT NULL", c.Field))
	}
	return db
}

// ApplyMongo 应用MongoDB查询条件
func (c Condition) ApplyMongo(_ *options.FindOptions, filter *bson.D) {
	switch strings.ToUpper(c.Operator) {
	case OpEQ:
		*filter = append(*filter, bson.E{Key: c.Field, Value: c.Value})
	case OpNE:
		*filter = append(*filter, bson.E{Key: c.Field, Value: bson.M{"$ne": c.Value}})
	case OpGT:
		*filter = append(*filter, bson.E{Key: c.Field, Value: bson.M{"$gt": c.Value}})
	case OpGTE:
		*filter = append(*filter, bson.E{Key: c.Field, Value: bson.M{"$gte": c.Value}})
	case OpLT:
		*filter = append(*filter, bson.E{Key: c.Field, Value: bson.M{"$lt": c.Value}})
	case OpLTE:
		*filter = append(*filter, bson.E{Key: c.Field, Value: bson.M{"$lte": c.Value}})
	case OpLIKE:
		*filter = append(*filter, bson.E{Key: c.Field, Value: bson.M{"$regex": c.Value}})
	case OpIN:
		*filter = append(*filter, bson.E{Key: c.Field, Value: bson.M{"$in": c.Value}})
	case OpNOTIN:
		*filter = append(*filter, bson.E{Key: c.Field, Value: bson.M{"$nin": c.Value}})
	case OpBETWEEN:
		if values, ok := c.Value.([]interface{}); ok && len(values) == 2 {
			*filter = append(*filter, bson.E{Key: c.Field, Value: bson.M{
				"$gte": values[0],
				"$lte": values[1],
			}})
		}
	case OpNULL:
		*filter = append(*filter, bson.E{Key: c.Field, Value: nil})
	case OpNOTNULL:
		*filter = append(*filter, bson.E{Key: c.Field, Value: bson.M{"$ne": nil}})
	}
}

// Order 排序条件
type Order struct {
	Field string // 排序字段
	Desc  bool   // 是否降序
}

// Apply 应用MySQL排序
func (o Order) Apply(db *gorm.DB) *gorm.DB {
	direction := "ASC"
	if o.Desc {
		direction = "DESC"
	}
	return db.Order(fmt.Sprintf("%s %s", o.Field, direction))
}

// ApplyMongo 应用MongoDB排序
func (o Order) ApplyMongo(opts *options.FindOptions, _ *bson.D) {
	direction := 1
	if o.Desc {
		direction = -1
	}
	if opts.Sort == nil {
		opts.Sort = bson.D{}
	}
	opts.Sort = append(opts.Sort.(bson.D), bson.E{Key: o.Field, Value: direction})
}

// Select 字段选择
type Select struct {
	Fields []string // 选择的字段列表
}

// Apply 应用MySQL字段选择
func (s Select) Apply(db *gorm.DB) *gorm.DB {
	return db.Select(s.Fields)
}

// ApplyMongo 应用MongoDB字段选择
func (s Select) ApplyMongo(opts *options.FindOptions, _ *bson.D) {
	projection := bson.D{}
	for _, field := range s.Fields {
		projection = append(projection, bson.E{Key: field, Value: 1})
	}
	opts.Projection = projection
}

// BuildQuery 构建查询条件
func BuildQuery(builders ...QueryBuilder) QueryBuilder {
	return &compositeBuilder{builders: builders}
}

// compositeBuilder 组合查询构建器
type compositeBuilder struct {
	builders []QueryBuilder
}

func (c *compositeBuilder) Apply(db *gorm.DB) *gorm.DB {
	for _, builder := range c.builders {
		db = builder.Apply(db)
	}
	return db
}

func (c *compositeBuilder) ApplyMongo(opts *options.FindOptions, filter *bson.D) {
	for _, builder := range c.builders {
		builder.ApplyMongo(opts, filter)
	}
}
