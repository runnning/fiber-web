package query

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

// ===== MongoDB查询构建器 =====

// MongoQuery MongoDB查询结构
type MongoQuery struct {
	Filter      bson.M             // 过滤条件
	Projection  bson.M             // 字段投影
	Collation   *options.Collation // 排序规则
	conditions  []Condition        // 条件列表
	fields      []string           // 查询字段
	sorts       bson.D             // 排序
	limit       int64              // 限制
	skip        int64              // 跳过
	pipeline    []bson.D           // 聚合管道
	isAggregate bool               // 是否使用聚合查询
}

// NewMongoQuery 创建新的MongoDB查询
func NewMongoQuery() *MongoQuery {
	return &MongoQuery{
		Filter:     bson.M{},
		Projection: bson.M{},
		conditions: make([]Condition, 0),
		fields:     make([]string, 0),
		sorts:      make(bson.D, 0),
		pipeline:   make([]bson.D, 0),
	}
}

// Where 添加条件
func (q *MongoQuery) Where(condition Condition) QueryBuilder {
	if condition != nil {
		q.conditions = append(q.conditions, condition)
	}
	return q
}

// WhereSimple 添加简单条件
func (q *MongoQuery) WhereSimple(field string, op Operator, value interface{}) QueryBuilder {
	return q.Where(NewSimpleCondition(field, op, value))
}

// WhereIn 添加IN条件
func (q *MongoQuery) WhereIn(field string, values []interface{}) QueryBuilder {
	return q.Where(NewSimpleCondition(field, OpIn, values))
}

// WhereGroup 添加条件组
func (q *MongoQuery) WhereGroup(logic LogicOperator, conditions ...Condition) QueryBuilder {
	if len(conditions) > 0 {
		return q.Where(NewGroupCondition(logic, conditions...))
	}
	return q
}

// WhereRaw 添加原始条件
func (q *MongoQuery) WhereRaw(raw interface{}) QueryBuilder {
	if raw != nil {
		if filter, ok := raw.(bson.M); ok {
			for k, v := range filter {
				q.Filter[k] = v
			}
		}
	}
	return q
}

// Select 设置查询字段
func (q *MongoQuery) Select(fields ...string) QueryBuilder {
	q.fields = append(q.fields, fields...)
	for _, field := range fields {
		q.Projection[field] = 1
	}
	return q
}

// OrderBy 设置排序
func (q *MongoQuery) OrderBy(field string, direction string) QueryBuilder {
	order := 1 // 默认升序
	if strings.ToUpper(direction) == "DESC" {
		order = -1
	}
	q.sorts = append(q.sorts, bson.E{Key: field, Value: order})
	return q
}

// GroupBy 设置分组
func (q *MongoQuery) GroupBy(field string) QueryBuilder {
	if field == "" {
		return q
	}

	// 标记为聚合查询
	q.isAggregate = true

	// 构建$group stage
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$" + field},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}},
	}

	// 将分组阶段添加到管道中
	q.pipeline = append(q.pipeline, groupStage)

	return q
}

// Having 设置分组条件
func (q *MongoQuery) Having(condition Condition) QueryBuilder {
	if condition == nil {
		return q
	}

	// 标记为聚合查询
	q.isAggregate = true

	// 构建$match stage用于Having条件
	matchStage := bson.D{{Key: "$match", Value: q.buildHavingCondition(condition)}}

	// 将匹配阶段添加到管道中
	q.pipeline = append(q.pipeline, matchStage)

	return q
}

// Limit 设置限制
func (q *MongoQuery) Limit(limit int) QueryBuilder {
	q.limit = int64(limit)
	return q
}

// Offset 设置偏移
func (q *MongoQuery) Offset(offset int) QueryBuilder {
	q.skip = int64(offset)
	return q
}

// Join 添加连接查询
func (q *MongoQuery) Join(table string, condition string) QueryBuilder {
	if table == "" || condition == "" {
		return q
	}

	// 标记为聚合查询
	q.isAggregate = true

	// 解析连接条件
	parts := strings.Split(condition, "=")
	if len(parts) != 2 {
		return q
	}

	localField := strings.TrimSpace(parts[0])
	foreignField := strings.TrimSpace(parts[1])

	// 移除字段名中的表名前缀
	localField = strings.TrimPrefix(localField, table+".")

	// 构建$lookup stage
	lookupStage := bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: table},
			{Key: "localField", Value: localField},
			{Key: "foreignField", Value: foreignField},
			{Key: "as", Value: table},
		}},
	}

	// 将连接阶段添加到管道中
	q.pipeline = append(q.pipeline, lookupStage)

	// 添加展开阶段，将数组转换为文档
	unwindStage := bson.D{
		{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$" + table},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}},
	}

	q.pipeline = append(q.pipeline, unwindStage)

	return q
}

// Build 实现QueryBuilder接口
func (q *MongoQuery) Build() interface{} {
	// 如果是聚合查询，返回聚合管道
	if q.isAggregate {
		// 添加初始的$match stage（如果有过滤条件）
		if len(q.Filter) > 0 {
			matchStage := bson.D{{Key: "$match", Value: q.Filter}}
			q.pipeline = append([]bson.D{matchStage}, q.pipeline...)
		}

		// 添加排序stage（如果有排序）
		if len(q.sorts) > 0 {
			sortStage := bson.D{{Key: "$sort", Value: q.sorts}}
			q.pipeline = append(q.pipeline, sortStage)
		}

		// 添加限制和跳过stage
		if q.limit > 0 {
			limitStage := bson.D{{Key: "$limit", Value: q.limit}}
			q.pipeline = append(q.pipeline, limitStage)
		}

		if q.skip > 0 {
			skipStage := bson.D{{Key: "$skip", Value: q.skip}}
			q.pipeline = append(q.pipeline, skipStage)
		}

		return q.pipeline
	}

	// 处理条件
	q.buildConditions()

	return q.Filter
}

// buildConditions 构建条件
func (q *MongoQuery) buildConditions() {
	for _, condition := range q.conditions {
		q.buildCondition(condition)
	}
}

// buildCondition 构建单个条件
func (q *MongoQuery) buildCondition(condition Condition) {
	if condition == nil {
		return
	}

	switch condition.GetType() {
	case ConditionTypeSimple:
		q.buildSimpleCondition(condition.(*SimpleCondition))
	case ConditionTypeGroup:
		q.buildGroupCondition(condition.(*GroupCondition))
	case ConditionTypeRaw:
		q.buildRawCondition(condition.(*RawCondition))
	}
}

// buildSimpleCondition 构建简单条件
func (q *MongoQuery) buildSimpleCondition(condition *SimpleCondition) {
	field := condition.Field
	value := condition.Value

	switch condition.Operator {
	case OpEq:
		q.Filter[field] = value
	case OpNe:
		q.Filter[field] = bson.M{"$ne": value}
	case OpGt:
		q.Filter[field] = bson.M{"$gt": value}
	case OpGte:
		q.Filter[field] = bson.M{"$gte": value}
	case OpLt:
		q.Filter[field] = bson.M{"$lt": value}
	case OpLte:
		q.Filter[field] = bson.M{"$lte": value}
	case OpIn:
		q.Filter[field] = bson.M{"$in": value}
	case OpNin:
		q.Filter[field] = bson.M{"$nin": value}
	case OpContains:
		if strVal, ok := value.(string); ok {
			q.Filter[field] = bson.M{"$regex": strVal, "$options": "i"}
		}
	case OpStartsWith:
		if strVal, ok := value.(string); ok {
			q.Filter[field] = bson.M{"$regex": "^" + strVal, "$options": "i"}
		}
	case OpEndsWith:
		if strVal, ok := value.(string); ok {
			q.Filter[field] = bson.M{"$regex": strVal + "$", "$options": "i"}
		}
	case OpExists:
		if boolVal, ok := value.(bool); ok {
			q.Filter[field] = bson.M{"$exists": boolVal}
		}
	case OpBetween:
		if values, ok := value.([]interface{}); ok && len(values) == 2 {
			q.Filter[field] = bson.M{
				"$gte": values[0],
				"$lte": values[1],
			}
		}
	case OpIsNull:
		q.Filter[field] = bson.M{"$eq": nil}
	case OpNotNull:
		q.Filter[field] = bson.M{"$ne": nil}
	}
}

// buildGroupCondition 构建条件组
func (q *MongoQuery) buildGroupCondition(condition *GroupCondition) {
	if len(condition.Conditions) == 0 {
		return
	}

	var operator string
	switch condition.Logic {
	case LogicAnd:
		operator = "$and"
	case LogicOr:
		operator = "$or"
	default:
		return
	}

	subConditions := make([]bson.M, 0, len(condition.Conditions))

	for _, subCondition := range condition.Conditions {
		subQuery := NewMongoQuery()
		subQuery.buildCondition(subCondition)
		if len(subQuery.Filter) > 0 {
			subConditions = append(subConditions, subQuery.Filter)
		}
	}

	if len(subConditions) > 0 {
		if existing, ok := q.Filter[operator]; ok {
			if existingArray, ok := existing.([]bson.M); ok {
				q.Filter[operator] = append(existingArray, subConditions...)
			} else {
				q.Filter[operator] = subConditions
			}
		} else {
			q.Filter[operator] = subConditions
		}
	}
}

// buildRawCondition 构建原始条件
func (q *MongoQuery) buildRawCondition(condition *RawCondition) {
	if raw, ok := condition.Raw.(bson.M); ok {
		for k, v := range raw {
			q.Filter[k] = v
		}
	}
}

// GetFindOptions 获取查询选项
func (q *MongoQuery) GetFindOptions() *options.FindOptions {
	opts := options.Find()

	// 设置投影
	if len(q.Projection) > 0 {
		opts.SetProjection(q.Projection)
	}

	// 设置排序
	if len(q.sorts) > 0 {
		opts.SetSort(q.sorts)
	}

	// 设置限制
	if q.limit > 0 {
		opts.SetLimit(q.limit)
	}

	// 设置偏移
	if q.skip > 0 {
		opts.SetSkip(q.skip)
	}

	// 设置排序规则
	if q.Collation != nil {
		opts.SetCollation(q.Collation)
	}

	return opts
}

// buildHavingCondition 构建Having条件
func (q *MongoQuery) buildHavingCondition(condition Condition) bson.M {
	switch condition.GetType() {
	case ConditionTypeSimple:
		return q.buildSimpleHavingCondition(condition.(*SimpleCondition))
	case ConditionTypeGroup:
		return q.buildGroupHavingCondition(condition.(*GroupCondition))
	case ConditionTypeRaw:
		if raw, ok := condition.(*RawCondition).Raw.(bson.M); ok {
			return raw
		}
	}
	return bson.M{}
}

// buildSimpleHavingCondition 构建简单Having条件
func (q *MongoQuery) buildSimpleHavingCondition(condition *SimpleCondition) bson.M {
	field := condition.Field
	value := condition.Value

	switch condition.Operator {
	case OpEq:
		return bson.M{field: value}
	case OpGt:
		return bson.M{field: bson.M{"$gt": value}}
	case OpGte:
		return bson.M{field: bson.M{"$gte": value}}
	case OpLt:
		return bson.M{field: bson.M{"$lt": value}}
	case OpLte:
		return bson.M{field: bson.M{"$lte": value}}
	default:
		return bson.M{}
	}
}

// buildGroupHavingCondition 构建分组Having条件
func (q *MongoQuery) buildGroupHavingCondition(condition *GroupCondition) bson.M {
	if len(condition.Conditions) == 0 {
		return bson.M{}
	}

	conditions := make([]bson.M, 0)
	for _, cond := range condition.Conditions {
		conditions = append(conditions, q.buildHavingCondition(cond))
	}

	op := "$and"
	if condition.Logic == LogicOr {
		op = "$or"
	}

	return bson.M{op: conditions}
}

// ===== MongoDB查询工厂 =====

// MongoQueryFactory MongoDB查询工厂
type MongoQueryFactory struct{}

// NewMongoQueryFactory 创建MongoDB查询工厂
func NewMongoQueryFactory() *MongoQueryFactory {
	return &MongoQueryFactory{}
}

// NewQuery 创建新的查询构建器
func (f *MongoQueryFactory) NewQuery() QueryBuilder {
	return NewMongoQuery()
}

// ===== MySQL查询构建器 =====

// JoinType 连接类型
type JoinType string

const (
	JoinTypeInner JoinType = "INNER JOIN"
	JoinTypeLeft  JoinType = "LEFT JOIN"
	JoinTypeRight JoinType = "RIGHT JOIN"
	JoinTypeCross JoinType = "CROSS JOIN"
)

// MySQLQuery MySQL查询结构
type MySQLQuery struct {
	db         *gorm.DB    // 原始数据库连接
	conditions []Condition // 查询条件
	fields     []string    // 查询字段
	joins      []JoinInfo  // 连接查询
	groupBy    []string    // 分组字段
	having     []Condition // 分组条件
	limit      int         // 限制
	offset     int         // 偏移
	orders     []string    // 排序
}

// JoinInfo 连接信息
type JoinInfo struct {
	Type      JoinType // 连接类型
	Table     string   // 表名
	Condition string   // 连接条件
}

// NewMySQLQuery 创建新的MySQL查询
func NewMySQLQuery(db *gorm.DB) *MySQLQuery {
	return &MySQLQuery{
		db:         db,
		conditions: make([]Condition, 0),
		fields:     make([]string, 0),
		joins:      make([]JoinInfo, 0),
		groupBy:    make([]string, 0),
		having:     make([]Condition, 0),
		orders:     make([]string, 0),
	}
}

// Where 添加条件
func (q *MySQLQuery) Where(condition Condition) QueryBuilder {
	if condition != nil {
		q.conditions = append(q.conditions, condition)
	}
	return q
}

// WhereSimple 添加简单条件
func (q *MySQLQuery) WhereSimple(field string, op Operator, value interface{}) QueryBuilder {
	return q.Where(NewSimpleCondition(field, op, value))
}

// WhereIn 添加IN条件
func (q *MySQLQuery) WhereIn(field string, values []interface{}) QueryBuilder {
	return q.Where(NewSimpleCondition(field, OpIn, values))
}

// WhereGroup 添加条件组
func (q *MySQLQuery) WhereGroup(logic LogicOperator, conditions ...Condition) QueryBuilder {
	if len(conditions) > 0 {
		return q.Where(NewGroupCondition(logic, conditions...))
	}
	return q
}

// WhereRaw 添加原始条件
func (q *MySQLQuery) WhereRaw(raw interface{}) QueryBuilder {
	if raw != nil {
		if db, ok := raw.(*gorm.DB); ok {
			q.db = db
		}
	}
	return q
}

// Select 设置查询字段
func (q *MySQLQuery) Select(fields ...string) QueryBuilder {
	q.fields = append(q.fields, fields...)
	return q
}

// OrderBy 设置排序
func (q *MySQLQuery) OrderBy(field string, direction string) QueryBuilder {
	if field != "" {
		order := field
		if direction != "" {
			order += " " + direction
		}
		q.orders = append(q.orders, order)
	}
	return q
}

// GroupBy 设置分组
func (q *MySQLQuery) GroupBy(field string) QueryBuilder {
	if field != "" {
		q.groupBy = append(q.groupBy, field)
	}
	return q
}

// Having 设置分组条件
func (q *MySQLQuery) Having(condition Condition) QueryBuilder {
	if condition != nil {
		q.having = append(q.having, condition)
	}
	return q
}

// Limit 设置限制
func (q *MySQLQuery) Limit(limit int) QueryBuilder {
	q.limit = limit
	return q
}

// Offset 设置偏移
func (q *MySQLQuery) Offset(offset int) QueryBuilder {
	q.offset = offset
	return q
}

// Join 添加连接
func (q *MySQLQuery) Join(table string, condition string) QueryBuilder {
	return q.JoinWithType(table, condition, JoinTypeInner)
}

// JoinWithType 使用指定类型添加连接
func (q *MySQLQuery) JoinWithType(table string, condition string, joinType JoinType) QueryBuilder {
	if table != "" && condition != "" {
		q.joins = append(q.joins, JoinInfo{
			Type:      joinType,
			Table:     table,
			Condition: condition,
		})
	}
	return q
}

// LeftJoin 添加左连接
func (q *MySQLQuery) LeftJoin(table string, condition string) QueryBuilder {
	return q.JoinWithType(table, condition, JoinTypeLeft)
}

// RightJoin 添加右连接
func (q *MySQLQuery) RightJoin(table string, condition string) QueryBuilder {
	return q.JoinWithType(table, condition, JoinTypeRight)
}

// CrossJoin 添加交叉连接
func (q *MySQLQuery) CrossJoin(table string, condition string) QueryBuilder {
	return q.JoinWithType(table, condition, JoinTypeCross)
}

// Build 实现QueryBuilder接口
func (q *MySQLQuery) Build() interface{} {
	return q.buildQuery()
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
		joinStr := fmt.Sprintf("%s %s ON %s", join.Type, join.Table, join.Condition)
		query = query.Joins(joinStr)
	}

	// 添加条件
	query = q.buildConditions(query)

	// 添加排序
	for _, order := range q.orders {
		query = query.Order(order)
	}

	// 添加分组
	if len(q.groupBy) > 0 {
		query = query.Group(strings.Join(q.groupBy, ", "))

		// 添加分组条件
		if len(q.having) > 0 {
			query = q.buildHavingConditions(query)
		}
	}

	// 添加限制和偏移
	if q.limit > 0 {
		query = query.Limit(q.limit)
	}

	if q.offset > 0 {
		query = query.Offset(q.offset)
	}

	return query
}

// buildConditions 构建所有条件
func (q *MySQLQuery) buildConditions(query *gorm.DB) *gorm.DB {
	for _, condition := range q.conditions {
		query = q.buildCondition(query, condition)
	}
	return query
}

// buildCondition 构建单个条件
func (q *MySQLQuery) buildCondition(query *gorm.DB, condition Condition) *gorm.DB {
	if condition == nil {
		return query
	}

	switch condition.GetType() {
	case ConditionTypeSimple:
		return q.buildSimpleCondition(query, condition.(*SimpleCondition))
	case ConditionTypeGroup:
		return q.buildGroupCondition(query, condition.(*GroupCondition))
	case ConditionTypeRaw:
		return q.buildRawCondition(query, condition.(*RawCondition))
	default:
		return query
	}
}

// buildSimpleCondition 构建简单条件
func (q *MySQLQuery) buildSimpleCondition(query *gorm.DB, condition *SimpleCondition) *gorm.DB {
	field := condition.Field
	value := condition.Value

	switch condition.Operator {
	case OpEq:
		return query.Where(fmt.Sprintf("%s = ?", field), value)
	case OpNe:
		return query.Where(fmt.Sprintf("%s != ?", field), value)
	case OpGt:
		return query.Where(fmt.Sprintf("%s > ?", field), value)
	case OpGte:
		return query.Where(fmt.Sprintf("%s >= ?", field), value)
	case OpLt:
		return query.Where(fmt.Sprintf("%s < ?", field), value)
	case OpLte:
		return query.Where(fmt.Sprintf("%s <= ?", field), value)
	case OpIn:
		return query.Where(fmt.Sprintf("%s IN ?", field), value)
	case OpNin:
		return query.Where(fmt.Sprintf("%s NOT IN ?", field), value)
	case OpContains:
		if strVal, ok := value.(string); ok {
			return query.Where(fmt.Sprintf("%s LIKE ?", field), "%"+strVal+"%")
		}
	case OpStartsWith:
		if strVal, ok := value.(string); ok {
			return query.Where(fmt.Sprintf("%s LIKE ?", field), strVal+"%")
		}
	case OpEndsWith:
		if strVal, ok := value.(string); ok {
			return query.Where(fmt.Sprintf("%s LIKE ?", field), "%"+strVal)
		}
	case OpExists:
		if boolVal, ok := value.(bool); ok {
			if boolVal {
				return query.Where(fmt.Sprintf("%s IS NOT NULL", field))
			} else {
				return query.Where(fmt.Sprintf("%s IS NULL", field))
			}
		}
	case OpBetween:
		if values, ok := value.([]interface{}); ok && len(values) == 2 {
			return query.Where(fmt.Sprintf("%s BETWEEN ? AND ?", field), values[0], values[1])
		}
	case OpIsNull:
		return query.Where(fmt.Sprintf("%s IS NULL", field))
	case OpNotNull:
		return query.Where(fmt.Sprintf("%s IS NOT NULL", field))
	}

	return query
}

// buildGroupCondition 构建条件组
func (q *MySQLQuery) buildGroupCondition(query *gorm.DB, condition *GroupCondition) *gorm.DB {
	if len(condition.Conditions) == 0 {
		return query
	}

	switch condition.Logic {
	case LogicAnd:
		// 对于AND条件，直接链式调用where即可
		for _, subCondition := range condition.Conditions {
			query = q.buildCondition(query, subCondition)
		}
		return query
	case LogicOr:
		// 对于OR条件，需要使用Or方法
		return query.Where(func(db *gorm.DB) *gorm.DB {
			for i, subCondition := range condition.Conditions {
				if i == 0 {
					db = q.buildCondition(db, subCondition)
				} else {
					db = db.Or(func(subDb *gorm.DB) *gorm.DB {
						return q.buildCondition(subDb, subCondition)
					})
				}
			}
			return db
		})
	default:
		return query
	}
}

// buildRawCondition 构建原始条件
func (q *MySQLQuery) buildRawCondition(query *gorm.DB, condition *RawCondition) *gorm.DB {
	if condition.Raw != nil {
		if db, ok := condition.Raw.(*gorm.DB); ok {
			return db
		}
	}
	return query
}

// buildHavingConditions 构建所有Having条件
func (q *MySQLQuery) buildHavingConditions(query *gorm.DB) *gorm.DB {
	for _, condition := range q.having {
		query = q.buildHavingCondition(query, condition)
	}
	return query
}

// buildHavingCondition 构建Having条件
func (q *MySQLQuery) buildHavingCondition(query *gorm.DB, condition Condition) *gorm.DB {
	if condition == nil {
		return query
	}

	switch condition.GetType() {
	case ConditionTypeSimple:
		return q.buildSimpleHavingCondition(query, condition.(*SimpleCondition))
	case ConditionTypeGroup:
		return q.buildGroupHavingCondition(query, condition.(*GroupCondition))
	case ConditionTypeRaw:
		return q.buildRawHavingCondition(query, condition.(*RawCondition))
	}
	return query
}

// buildSimpleHavingCondition 构建简单Having条件
func (q *MySQLQuery) buildSimpleHavingCondition(query *gorm.DB, condition *SimpleCondition) *gorm.DB {
	field := condition.Field
	value := condition.Value

	switch condition.Operator {
	case OpEq:
		return query.Having(fmt.Sprintf("%s = ?", field), value)
	case OpNe:
		return query.Having(fmt.Sprintf("%s != ?", field), value)
	case OpGt:
		return query.Having(fmt.Sprintf("%s > ?", field), value)
	case OpGte:
		return query.Having(fmt.Sprintf("%s >= ?", field), value)
	case OpLt:
		return query.Having(fmt.Sprintf("%s < ?", field), value)
	case OpLte:
		return query.Having(fmt.Sprintf("%s <= ?", field), value)
	case OpIn:
		return query.Having(fmt.Sprintf("%s IN ?", field), value)
	case OpNin:
		return query.Having(fmt.Sprintf("%s NOT IN ?", field), value)
	case OpBetween:
		if values, ok := value.([]interface{}); ok && len(values) == 2 {
			return query.Having(fmt.Sprintf("%s BETWEEN ? AND ?", field), values[0], values[1])
		}
	}
	return query
}

// buildGroupHavingCondition 构建Having条件组
func (q *MySQLQuery) buildGroupHavingCondition(query *gorm.DB, condition *GroupCondition) *gorm.DB {
	if len(condition.Conditions) == 0 {
		return query
	}

	return query.Having(func(db *gorm.DB) *gorm.DB {
		for i, cond := range condition.Conditions {
			if i == 0 || condition.Logic == LogicAnd {
				db = q.buildHavingCondition(db, cond)
			} else {
				db = db.Or(func(subDb *gorm.DB) *gorm.DB {
					return q.buildHavingCondition(subDb, cond)
				})
			}
		}
		return db
	})
}

// buildRawHavingCondition 构建原始Having条件
func (q *MySQLQuery) buildRawHavingCondition(query *gorm.DB, condition *RawCondition) *gorm.DB {
	if raw, ok := condition.Raw.(string); ok {
		return query.Having(raw)
	}
	return query
}

// ===== MySQL查询工厂 =====

// MySQLQueryFactory MySQL查询工厂
type MySQLQueryFactory struct {
	DB *gorm.DB
}

// NewMySQLQueryFactory 创建MySQL查询工厂
func NewMySQLQueryFactory(db *gorm.DB) *MySQLQueryFactory {
	return &MySQLQueryFactory{
		DB: db,
	}
}

// NewQuery 创建新的查询构建器
func (f *MySQLQueryFactory) NewQuery() QueryBuilder {
	return NewMySQLQuery(f.DB)
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
