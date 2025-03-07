# Query Package

这是一个用于MySQL和MongoDB的通用分页条件查询包。它提供了简单且统一的API来处理分页和条件查询，支持泛型。

## 特性

- 支持MySQL（基于GORM）和MongoDB
- 统一的查询条件接口
- 泛型支持，提供类型安全
- 灵活的操作符支持
- 支持排序和字段选择
- 自动处理分页逻辑
- 链式调用API
- 可扩展的设计

## 安装

```bash
go get -u github.com/yourusername/query
```

## 快速开始

### 基本使用

```go
// 创建查询条件
query := query.NewQuery().
    SetPage(1, 10).
    AddCondition("age", query.OpGte, 18).
    AddOrderBy("created_at DESC").
    Select("id", "name", "age")

// MySQL查询
type User struct {
    ID   uint   `gorm:"primarykey"`
    Name string
    Age  int
}

mysqlQuerier := query.NewMySQLQuerier[User](db)
result, err := mysqlQuerier.FindPage(query)

// MongoDB查询
type Product struct {
    ID    primitive.ObjectID `bson:"_id"`
    Name  string            `bson:"name"`
    Price float64           `bson:"price"`
}

mongoQuerier := query.NewMongoQuerier[Product](collection)
result, err := mongoQuerier.FindPage(ctx, query)
```

### 高级用法

```go
// 1. 禁用分页，查询所有数据
query := query.NewQuery().
    DisablePagination().
    AddCondition("status", query.OpEq, 1)

// 2. 多条件查询
query := query.NewQuery().
    SetPage(1, 20).
    AddCondition("age", query.OpGte, 18).
    AddCondition("status", query.OpEq, 1).
    AddCondition("name", query.OpLike, "张").
    AddOrderBy("age DESC").
    AddOrderBy("created_at DESC")

// 3. 字段选择
query := query.NewQuery().
    Select("id", "name", "age"). // 设置要查询的字段
    AddSelect("status")          // 追加字段

// 4. 区间查询
query := query.NewQuery().
    AddCondition("age", query.OpBetween, []interface{}{18, 30}).
    AddCondition("created_at", query.OpBetween, []interface{}{startTime, endTime})

// 5. IN查询
query := query.NewQuery().
    AddCondition("status", query.OpIn, []int{1, 2, 3}).
    AddCondition("type", query.OpNotIn, []string{"deleted", "disabled"})
```

## 默认值

- 分页：默认启用
- 页码（Page）：默认为 1
- 每页数量（PageSize）：默认为 10
- 切片字段预分配容量：
  - Conditions: 4
  - OrderBy: 2
  - SelectFields: 4

## 支持的操作符

| 操作符 | 说明 | 示例 |
|--------|------|------|
| OpEq | 等于 | `AddCondition("status", OpEq, 1)` |
| OpNe | 不等于 | `AddCondition("status", OpNe, 0)` |
| OpGt | 大于 | `AddCondition("age", OpGt, 18)` |
| OpGte | 大于等于 | `AddCondition("age", OpGte, 18)` |
| OpLt | 小于 | `AddCondition("price", OpLt, 100)` |
| OpLte | 小于等于 | `AddCondition("price", OpLte, 100)` |
| OpIn | 在列表中 | `AddCondition("status", OpIn, []int{1,2})` |
| OpNotIn | 不在列表中 | `AddCondition("status", OpNotIn, []int{0,-1})` |
| OpLike | 模糊匹配 | `AddCondition("name", OpLike, "张")` |
| OpNotLike | 不匹配 | `AddCondition("name", OpNotLike, "李")` |
| OpBetween | 区间 | `AddCondition("age", OpBetween, []interface{}{18, 30})` |
| OpNotBetween | 不在区间 | `AddCondition("age", OpNotBetween, []interface{}{0, 18})` |

## API参考

### Query 方法

```go
// 创建新查询
NewQuery() *Query

// 分页相关
SetPage(page, pageSize int) *Query
SetPagination(pagination *Pagination) *Query
DisablePagination() *Query
EnablePaginationFunc() *Query

// 字段选择
Select(fields ...string) *Query
AddSelect(fields ...string) *Query

// 条件和排序
AddCondition(field string, operator Operator, value interface{}) *Query
AddOrderBy(order string) *Query
```

### 查询器接口

```go
type Querier[T any] interface {
    FindPage(q *Query) (*PageResult[T], error)
}
```

## 注意事项

1. 需要Go 1.18或更高版本（因为使用了泛型特性）
2. MySQL查询器基于GORM，需要确保模型符合GORM的规范
3. MongoDB查询器使用官方的mongo-driver，需要确保模型有正确的bson标签
4. 排序字段格式为："字段名 ASC"或"字段名 DESC"，默认为ASC
5. MongoDB的FindPage方法需要额外的context.Context参数

## 最佳实践

1. 使用 `NewQuery()` 创建查询对象，避免手动初始化
2. 使用链式调用构建查询条件，提高代码可读性
3. 合理使用字段选择功能，避免查询不必要的字段
4. 需要查询全部数据时，使用 `DisablePagination()`
5. 对于大数据量查询，合理设置分页大小

## 贡献

欢迎提交Issue和Pull Request。

## 许可证

MIT License 