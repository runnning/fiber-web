package query

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoQuerier MongoDB查询器
type MongoQuerier[T any] struct {
	BaseQuerier[T]
	collection *mongo.Collection
}

// NewMongoQuerier 创建MongoDB查询器
func NewMongoQuerier[T any](collection *mongo.Collection) Querier[T] {
	if collection == nil {
		panic("collection cannot be nil")
	}
	return &MongoQuerier[T]{
		collection: collection,
	}
}

// buildFilter 构建MongoDB查询条件
func (q *MongoQuerier[T]) buildFilter(query *Query) bson.D {
	if query == nil {
		return bson.D{}
	}

	filter := bson.D{}
	for _, cond := range query.Conditions {
		if cond.Field == "" {
			continue
		}

		switch cond.Operator {
		case OpEq:
			filter = append(filter, bson.E{Key: cond.Field, Value: cond.Value})
		case OpNe:
			filter = append(filter, bson.E{Key: cond.Field, Value: bson.M{"$ne": cond.Value}})
		case OpGt:
			filter = append(filter, bson.E{Key: cond.Field, Value: bson.M{"$gt": cond.Value}})
		case OpGte:
			filter = append(filter, bson.E{Key: cond.Field, Value: bson.M{"$gte": cond.Value}})
		case OpLt:
			filter = append(filter, bson.E{Key: cond.Field, Value: bson.M{"$lt": cond.Value}})
		case OpLte:
			filter = append(filter, bson.E{Key: cond.Field, Value: bson.M{"$lte": cond.Value}})
		case OpIn:
			filter = append(filter, bson.E{Key: cond.Field, Value: bson.M{"$in": cond.Value}})
		case OpNotIn:
			filter = append(filter, bson.E{Key: cond.Field, Value: bson.M{"$nin": cond.Value}})
		case OpLike:
			if str, ok := cond.Value.(string); ok {
				filter = append(filter, bson.E{Key: cond.Field, Value: bson.M{"$regex": str, "$options": "i"}})
			}
		case OpNotLike:
			if str, ok := cond.Value.(string); ok {
				filter = append(filter, bson.E{Key: cond.Field, Value: bson.M{"$not": bson.M{"$regex": str, "$options": "i"}}})
			}
		case OpBetween:
			if values, ok := cond.Value.([]interface{}); ok && len(values) == 2 {
				filter = append(filter, bson.E{Key: cond.Field, Value: bson.M{"$gte": values[0], "$lte": values[1]}})
			}
		case OpNotBetween:
			if values, ok := cond.Value.([]interface{}); ok && len(values) == 2 {
				filter = append(filter, bson.E{Key: cond.Field, Value: bson.M{"$not": bson.M{"$gte": values[0], "$lte": values[1]}}})
			}
		}
	}

	return filter
}

// buildSort 构建MongoDB排序条件
func (q *MongoQuerier[T]) buildSort(query *Query) bson.D {
	if query == nil || len(query.OrderBy) == 0 {
		return bson.D{}
	}

	sort := bson.D{}
	for _, order := range query.OrderBy {
		if order == "" {
			continue
		}

		field := order
		value := 1 // 默认升序

		if len(order) > 5 && order[len(order)-5:] == " DESC" {
			field = order[:len(order)-5]
			value = -1
		} else if len(order) > 4 && order[len(order)-4:] == " ASC" {
			field = order[:len(order)-4]
		}

		sort = append(sort, bson.E{Key: field, Value: value})
	}

	return sort
}

// buildProjection 构建字段选择
func (q *MongoQuerier[T]) buildProjection(query *Query) bson.D {
	if query == nil || len(query.SelectFields) == 0 {
		return nil
	}

	projection := bson.D{}
	for _, field := range query.SelectFields {
		if field != "" {
			projection = append(projection, bson.E{Key: field, Value: 1})
		}
	}
	return projection
}

// FindPage 分页查询
func (q *MongoQuerier[T]) FindPage(ctx context.Context, query *Query) (*PageResult[T], error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}
	if query == nil {
		query = NewQuery()
	}

	filter := q.buildFilter(query)

	// 获取总数
	total, err := q.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("count documents error: %w", err)
	}

	// 处理分页参数
	page, pageSize := q.HandlePagination(query.Pagination)

	// 构建查询选项
	findOptions := options.Find()

	// 如果启用分页，设置分页参数
	if query.EnablePagination {
		findOptions.SetSkip(int64((page - 1) * pageSize))
		findOptions.SetLimit(int64(pageSize))
	}

	// 添加排序
	if len(query.OrderBy) > 0 {
		findOptions.SetSort(q.buildSort(query))
	}

	// 添加字段选择
	if projection := q.buildProjection(query); projection != nil {
		findOptions.SetProjection(projection)
	}

	// 执行查询
	cursor, err := q.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("find documents error: %w", err)
	}
	defer cursor.Close(ctx)

	// 解码结果
	var list []T
	if err := cursor.All(ctx, &list); err != nil {
		return nil, fmt.Errorf("decode documents error: %w", err)
	}

	return q.NewPageResult(list, total, page, pageSize), nil
}

// Count 获取总数
func (q *MongoQuerier[T]) Count(ctx context.Context, query *Query) (int64, error) {
	if ctx == nil {
		return 0, fmt.Errorf("context cannot be nil")
	}

	filter := q.buildFilter(query)
	count, err := q.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count documents error: %w", err)
	}

	return count, nil
}

// First 获取单条记录
func (q *MongoQuerier[T]) First(ctx context.Context, query *Query) (*T, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	filter := q.buildFilter(query)
	findOptions := options.FindOne()

	// 添加字段选择
	if projection := q.buildProjection(query); projection != nil {
		findOptions.SetProjection(projection)
	}

	// 添加排序
	if len(query.OrderBy) > 0 {
		findOptions.SetSort(q.buildSort(query))
	}

	var result T
	err := q.collection.FindOne(ctx, filter, findOptions).Decode(&result)
	if err != nil {
		//if err == mongo.ErrNoDocuments {
		//	return nil, nil
		//}
		return nil, fmt.Errorf("find one document error: %w", err)
	}

	return &result, nil
}

// Find 获取记录列表（不分页）
func (q *MongoQuerier[T]) Find(ctx context.Context, query *Query) ([]T, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	filter := q.buildFilter(query)
	findOptions := options.Find()

	// 添加字段选择
	if projection := q.buildProjection(query); projection != nil {
		findOptions.SetProjection(projection)
	}

	// 添加排序
	if len(query.OrderBy) > 0 {
		findOptions.SetSort(q.buildSort(query))
	}

	cursor, err := q.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("find documents error: %w", err)
	}
	defer cursor.Close(ctx)

	var list []T
	if err := cursor.All(ctx, &list); err != nil {
		return nil, fmt.Errorf("decode documents error: %w", err)
	}

	return list, nil
}
