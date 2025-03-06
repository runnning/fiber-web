package query

import (
	"context"
	"errors"

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
	filter, err := p.parseQuery(query)
	if err != nil {
		return 0, err
	}
	return p.Collection.CountDocuments(ctx, filter)
}

// Find 查询数据列表
func (p *MongoProvider[T]) Find(ctx context.Context, query interface{}, req *PageRequest, result *[]T) error {
	filter, err := p.parseQuery(query)
	if err != nil {
		return err
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

// FindOne 查询单条记录
func (p *MongoProvider[T]) FindOne(ctx context.Context, query interface{}, result *T) error {
	filter, err := p.parseQuery(query)
	if err != nil {
		return err
	}

	return p.Collection.FindOne(ctx, filter).Decode(result)
}

// Insert 插入记录
func (p *MongoProvider[T]) Insert(ctx context.Context, data *T) error {
	_, err := p.Collection.InsertOne(ctx, data)
	return err
}

// Update 更新记录
func (p *MongoProvider[T]) Update(ctx context.Context, query interface{}, data map[string]interface{}) error {
	filter, err := p.parseQuery(query)
	if err != nil {
		return err
	}

	update := bson.M{"$set": data}
	_, err = p.Collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete 删除记录
func (p *MongoProvider[T]) Delete(ctx context.Context, query interface{}) error {
	filter, err := p.parseQuery(query)
	if err != nil {
		return err
	}

	_, err = p.Collection.DeleteOne(ctx, filter)
	return err
}

// Transaction 事务操作
func (p *MongoProvider[T]) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	session, err := p.Collection.Database().Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	callback := func(sessionContext mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessionContext)
	}

	_, err = session.WithTransaction(ctx, callback)
	return err
}

// parseQuery 解析查询条件
func (p *MongoProvider[T]) parseQuery(query interface{}) (bson.M, error) {
	if query == nil {
		return bson.M{}, nil
	}

	switch q := query.(type) {
	case bson.M:
		return q, nil
	case bson.D:
		return bson.M{}, nil
	case *MongoQuery:
		return q.Filter, nil
	default:
		return nil, errors.New("unsupported query type for MongoDB")
	}
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
	db, err := p.parseQuery(query)
	if err != nil {
		return 0, err
	}

	// 创建新的会话以避免影响原始查询
	db = db.Session(&gorm.Session{})

	// 确保设置了正确的表
	var model T
	db = db.Model(&model)

	// 应用上下文
	db = db.WithContext(ctx)

	// 执行查询
	err = db.Count(&total).Error
	return total, err
}

// Find 查询数据列表
func (p *MySQLProvider[T]) Find(ctx context.Context, query interface{}, req *PageRequest, result *[]T) error {
	db, err := p.parseQuery(query)
	if err != nil {
		return err
	}

	// 创建新的会话以避免影响原始查询
	db = db.Session(&gorm.Session{})

	// 确保设置了正确的表
	var model T
	db = db.Model(&model)

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

// FindOne 查询单条记录
func (p *MySQLProvider[T]) FindOne(ctx context.Context, query interface{}, result *T) error {
	db, err := p.parseQuery(query)
	if err != nil {
		return err
	}

	return db.WithContext(ctx).First(result).Error
}

// Insert 插入记录
func (p *MySQLProvider[T]) Insert(ctx context.Context, data *T) error {
	return p.DB.WithContext(ctx).Create(data).Error
}

// Update 更新记录
func (p *MySQLProvider[T]) Update(ctx context.Context, query interface{}, data map[string]interface{}) error {
	db, err := p.parseQuery(query)
	if err != nil {
		return err
	}

	return db.WithContext(ctx).Updates(data).Error
}

// Delete 删除记录
func (p *MySQLProvider[T]) Delete(ctx context.Context, query interface{}) error {
	db, err := p.parseQuery(query)
	if err != nil {
		return err
	}

	var model T
	return db.WithContext(ctx).Delete(&model).Error
}

// Transaction 事务操作
func (p *MySQLProvider[T]) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return p.DB.Transaction(func(tx *gorm.DB) error {
		// 创建事务上下文
		txCtx := context.WithValue(ctx, "tx", tx)
		return fn(txCtx)
	})
}

// parseQuery 解析查询条件
func (p *MySQLProvider[T]) parseQuery(query interface{}) (*gorm.DB, error) {
	if query == nil {
		return p.DB, nil
	}

	switch q := query.(type) {
	case *gorm.DB:
		return q, nil
	default:
		return nil, errors.New("unsupported query type for MySQL")
	}
}
