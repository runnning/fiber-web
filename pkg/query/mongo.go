package query

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoQuery MongoDB查询结构
type MongoQuery struct {
	Filter     bson.M             // 过滤条件
	Projection bson.M             // 字段投影
	Collation  *options.Collation // 排序规则
}

// NewMongoQuery 创建新的MongoDB查询
func NewMongoQuery() *MongoQuery {
	return &MongoQuery{
		Filter:     bson.M{},
		Projection: bson.M{},
	}
}

// SetFilter 设置过滤条件
func (q *MongoQuery) SetFilter(filter bson.M) *MongoQuery {
	q.Filter = filter
	return q
}

// AddFilter 添加过滤条件
func (q *MongoQuery) AddFilter(key string, value interface{}) *MongoQuery {
	q.Filter[key] = value
	return q
}

// SetProjection 设置字段投影
func (q *MongoQuery) SetProjection(fields ...string) *MongoQuery {
	for _, field := range fields {
		q.Projection[field] = 1
	}
	return q
}

// ExcludeFields 排除字段
func (q *MongoQuery) ExcludeFields(fields ...string) *MongoQuery {
	for _, field := range fields {
		q.Projection[field] = 0
	}
	return q
}

// SetCollation 设置排序规则
func (q *MongoQuery) SetCollation(collation *options.Collation) *MongoQuery {
	q.Collation = collation
	return q
}

// MongoPaginate MongoDB分页查询
func MongoPaginate[T any](ctx context.Context, coll *mongo.Collection, query *MongoQuery, req *PageRequest, result *[]T) (*PageResponse[T], error) {
	// 计算总记录数
	total, err := coll.CountDocuments(ctx, query.Filter)
	if err != nil {
		return nil, err
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
	cursor, err := coll.Find(ctx, query.Filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 解码结果
	if err := cursor.All(ctx, result); err != nil {
		return nil, err
	}

	return NewPageResponse(*result, total, req.Page, req.PageSize), nil
}

// Example usage:
/*
func GetUserList(ctx context.Context, coll *mongo.Collection, req *PageRequest) (*PageResponse, error) {
    var users []User

    // 创建查询构建器
    filter := NewFilterBuilder().
        AddCondition("age", OpGte, 18).
        AddCondition("status", OpEq, "active").
        AddArrayCondition("roles", OpIn, []string{"user", "admin"}).
        Build()

    // 创建MongoDB查询
    query := NewMongoQuery().
        SetFilter(filter).
        SetProjection("name", "email", "age").
        SetCollation(&options.Collation{
            Locale:          "zh",
            Strength:        1,
            CaseLevel:       false,
            NumericOrdering: true,
        })

    return MongoPaginate(ctx, coll, query, req, &users)
}

// 支持时间范围查询
func GetUsersByTimeRange(ctx context.Context, coll *mongo.Collection, req *PageRequest, startTime, endTime *time.Time) (*PageResponse, error) {
    var users []User

    // 创建查询构建器
    fb := NewFilterBuilder()

    // 添加时间范围条件
    timeConditions := ParseTimeRange("createTime", startTime, endTime)
    for _, condition := range timeConditions {
        fb.AddCondition(condition.Field, condition.Operator, condition.Value)
    }

    // 创建MongoDB查询
    query := NewMongoQuery().
        SetFilter(fb.Build())

    return MongoPaginate(ctx, coll, query, req, &users)
}

// 支持多字段模糊搜索
func SearchUsers(ctx context.Context, coll *mongo.Collection, req *PageRequest, searchText string) (*PageResponse, error) {
    var users []User

    // 创建查询构建器
    fb := NewFilterBuilder()

    // 添加搜索条件
    searchConditions := ParseSearchText(searchText, []string{"name", "email", "description"})
    for _, condition := range searchConditions {
        fb.AddCondition(condition.Field, condition.Operator, condition.Value)
    }

    // 创建MongoDB查询
    query := NewMongoQuery().
        SetFilter(fb.Build())

    return MongoPaginate(ctx, coll, query, req, &users)
}
*/
