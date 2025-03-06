package query

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

// ===== MongoDB数据提供者 =====

// MongoProvider MongoDB数据提供者
type MongoProvider[T any] struct {
	Collection *mongo.Collection
}

// NewMongoProvider 创建MongoDB数据提供者
func NewMongoProvider[T any](collection *mongo.Collection) *MongoProvider[T] {
	return &MongoProvider[T]{
		Collection: collection,
	}
}

// Count 计算符合条件的记录总数
func (p *MongoProvider[T]) Count(ctx context.Context, query interface{}) (int64, error) {
	filter, ok := query.(bson.M)
	if !ok {
		filter = bson.M{}
	}
	return p.Collection.CountDocuments(ctx, filter)
}

// Find 查询数据列表
func (p *MongoProvider[T]) Find(ctx context.Context, query interface{}, req *PageRequest, result *[]T) error {
	filter, ok := query.(bson.M)
	if !ok {
		filter = bson.M{}
	}

	// 构建查询选项
	opts := options.Find()
	opts.SetSkip(int64(req.Offset()))
	opts.SetLimit(int64(req.PageSize))

	// 排序
	if req.OrderBy != "" {
		order := 1 // 默认升序
		if req.Order == "DESC" {
			order = -1
		}
		opts.SetSort(bson.D{{Key: req.OrderBy, Value: order}})
	}

	// 执行查询
	cursor, err := p.Collection.Find(ctx, filter, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	// 解码结果
	return cursor.All(ctx, result)
}

// ===== MySQL数据提供者 =====

// MySQLProvider MySQL数据提供者
type MySQLProvider[T any] struct {
	DB *gorm.DB
}

// NewMySQLProvider 创建MySQL数据提供者
func NewMySQLProvider[T any](db *gorm.DB) *MySQLProvider[T] {
	return &MySQLProvider[T]{
		DB: db,
	}
}

// Count 计算符合条件的记录总数
func (p *MySQLProvider[T]) Count(ctx context.Context, query interface{}) (int64, error) {
	var total int64
	db, ok := query.(*gorm.DB)
	if !ok {
		db = p.DB
	}

	err := db.WithContext(ctx).Count(&total).Error
	return total, err
}

// Find 查询数据列表
func (p *MySQLProvider[T]) Find(ctx context.Context, query interface{}, req *PageRequest, result *[]T) error {
	db, ok := query.(*gorm.DB)
	if !ok {
		db = p.DB
	}

	// 添加上下文
	db = db.WithContext(ctx)

	// 排序
	if req.OrderBy != "" {
		order := req.OrderBy
		if req.Order != "" {
			order += " " + req.Order
		}
		db = db.Order(order)
	}

	// 分页查询
	return db.Offset(req.Offset()).Limit(req.PageSize).Find(result).Error
}
