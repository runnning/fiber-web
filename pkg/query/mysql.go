package query

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// MySQLQuerier MySQL查询器
type MySQLQuerier[T any] struct {
	BaseQuerier[T]
	db *gorm.DB
}

// NewMySQLQuerier 创建MySQL查询器
func NewMySQLQuerier[T any](db *gorm.DB) Querier[T] {
	if db == nil {
		panic("db cannot be nil")
	}
	return &MySQLQuerier[T]{
		db: db,
	}
}

// buildConditions 构建查询条件
func (q *MySQLQuerier[T]) buildConditions(query *Query) *gorm.DB {
	if query == nil {
		return q.db
	}

	db := q.db

	// 添加选择字段
	if len(query.SelectFields) > 0 {
		db = db.Select(query.SelectFields)
	}

	// 添加查询条件
	for _, cond := range query.Conditions {
		if cond.Field == "" {
			continue
		}

		switch cond.Operator {
		case OpEq:
			db = db.Where(fmt.Sprintf("%s = ?", cond.Field), cond.Value)
		case OpNe:
			db = db.Where(fmt.Sprintf("%s != ?", cond.Field), cond.Value)
		case OpGt:
			db = db.Where(fmt.Sprintf("%s > ?", cond.Field), cond.Value)
		case OpGte:
			db = db.Where(fmt.Sprintf("%s >= ?", cond.Field), cond.Value)
		case OpLt:
			db = db.Where(fmt.Sprintf("%s < ?", cond.Field), cond.Value)
		case OpLte:
			db = db.Where(fmt.Sprintf("%s <= ?", cond.Field), cond.Value)
		case OpIn:
			db = db.Where(fmt.Sprintf("%s IN ?", cond.Field), cond.Value)
		case OpNotIn:
			db = db.Where(fmt.Sprintf("%s NOT IN ?", cond.Field), cond.Value)
		case OpLike:
			db = db.Where(fmt.Sprintf("%s LIKE ?", cond.Field), fmt.Sprintf("%%%v%%", cond.Value))
		case OpNotLike:
			db = db.Where(fmt.Sprintf("%s NOT LIKE ?", cond.Field), fmt.Sprintf("%%%v%%", cond.Value))
		case OpBetween:
			if values, ok := cond.Value.([]any); ok && len(values) == 2 {
				db = db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", cond.Field), values[0], values[1])
			}
		case OpNotBetween:
			if values, ok := cond.Value.([]any); ok && len(values) == 2 {
				db = db.Where(fmt.Sprintf("%s NOT BETWEEN ? AND ?", cond.Field), values[0], values[1])
			}
		}
	}

	// 添加排序
	if len(query.OrderBy) > 0 {
		for _, order := range query.OrderBy {
			if order != "" {
				db = db.Order(order)
			}
		}
	}

	return db
}

// Count 获取总数
func (q *MySQLQuerier[T]) Count(ctx context.Context, query *Query) (int64, error) {
	if ctx == nil {
		return 0, fmt.Errorf("context cannot be nil")
	}

	var total int64
	db := q.buildConditions(query).WithContext(ctx)
	if err := db.Model(new(T)).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count records error: %w", err)
	}
	return total, nil
}

// First 获取单条记录
func (q *MySQLQuerier[T]) First(ctx context.Context, query *Query) (*T, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	var result T
	db := q.buildConditions(query).WithContext(ctx)
	if err := db.First(&result).Error; err != nil {
		//if errors.Is(err, gorm.ErrRecordNotFound) {
		//	return nil, nil
		//}
		return nil, fmt.Errorf("find first record error: %w", err)
	}
	return &result, nil
}

// Find 获取记录列表（不分页）
func (q *MySQLQuerier[T]) Find(ctx context.Context, query *Query) ([]T, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	var list []T
	db := q.buildConditions(query).WithContext(ctx)
	if err := db.Find(&list).Error; err != nil {
		return nil, fmt.Errorf("find records error: %w", err)
	}
	return list, nil
}

// FindPage 分页查询
func (q *MySQLQuerier[T]) FindPage(ctx context.Context, query *Query) (*PageResult[T], error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}
	if query == nil {
		query = NewQuery()
	}

	// 获取总数
	total, err := q.Count(ctx, query)
	if err != nil {
		return nil, err
	}

	// 处理分页参数
	page, pageSize := q.HandlePagination(query.Pagination)

	// 如果禁用分页，直接查询
	if !query.EnablePagination {
		list, err := q.Find(ctx, query)
		if err != nil {
			return nil, err
		}
		return q.NewPageResult(list, total, page, pageSize), nil
	}

	// 分页查询
	var list []T
	db := q.buildConditions(query).WithContext(ctx)
	offset := (page - 1) * pageSize
	if err := db.Limit(pageSize).Offset(offset).Find(&list).Error; err != nil {
		return nil, fmt.Errorf("find records with pagination error: %w", err)
	}

	return q.NewPageResult(list, total, page, pageSize), nil
}
