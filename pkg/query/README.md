# 查询包 (Query Package)

这个包提供了一个灵活、统一的查询构建器和数据提供者接口，支持MongoDB和MySQL等多种数据库。

## 主要改进

1. **统一的查询接口**：MongoDB和MySQL使用相同的接口和API，使代码更加一致和可维护
2. **灵活的条件构建**：支持简单条件、条件组（AND/OR）和原始条件，可以构建复杂的查询
3. **链式调用API**：所有查询构建器方法都支持链式调用，使代码更加简洁
4. **工厂模式**：使用工厂模式创建查询构建器，降低耦合度
5. **完整的CRUD支持**：数据提供者接口支持完整的CRUD操作和事务处理
6. **类型安全**：使用泛型确保类型安全
7. **职责分离**：分页请求只负责分页和排序，查询条件由查询构建器负责，符合单一职责原则

## 使用示例

### MongoDB示例

```go
// 创建MongoDB查询工厂
factory := query.NewMongoQueryFactory()

// 创建查询构建器
builder := factory.NewQuery()

// 构建查询条件
builder.WhereSimple("age", query.OpGte, 18).
        WhereSimple("status", query.OpEq, "active").
        WhereGroup(query.LogicOr,
            query.NewContainsCondition("name", "张"),
            query.NewContainsCondition("name", "李"),
        )

// 创建分页请求
pageReq := query.NewPageRequest(1, 10)
pageReq.OrderBy = "createdAt"
pageReq.Order = "DESC"

// 创建数据提供者
provider := query.NewMongoProvider[User](collection)

// 执行查询
var users []User
pageRes, err := query.Paginate(ctx, builder, provider, pageReq, &users)
if err != nil {
    // 处理错误
}

// 使用查询结果
fmt.Printf("总记录数: %d\n", pageRes.Total)
for _, user := range pageRes.List {
    fmt.Printf("用户: %s\n", user.Name)
}
```

### MySQL示例

```go
// 创建MySQL查询工厂
factory := query.NewMySQLQueryFactory(db)

// 创建查询构建器
builder := factory.NewQuery()

// 构建查询条件
builder.WhereSimple("age", query.OpGte, 18).
        WhereSimple("status", query.OpEq, "active").
        WhereGroup(query.LogicOr,
            query.NewContainsCondition("name", "张"),
            query.NewContainsCondition("name", "李"),
        ).
        Join("departments", "users.dept_id = departments.id").
        Select("users.*", "departments.name as dept_name").
        GroupBy("users.dept_id").
        Having(query.NewGtCondition("COUNT(*)", 5))

// 创建分页请求
pageReq := query.NewPageRequest(1, 10)
pageReq.OrderBy = "users.created_at"
pageReq.Order = "DESC"

// 创建数据提供者
provider := query.NewMySQLProvider[User](db)

// 执行查询
var users []User
pageRes, err := query.Paginate(ctx, builder, provider, pageReq, &users)
if err != nil {
    // 处理错误
}

// 使用查询结果
fmt.Printf("总记录数: %d\n", pageRes.Total)
for _, user := range pageRes.List {
    fmt.Printf("用户: %s, 部门: %s\n", user.Name, user.DeptName)
}
```

### 高级查询示例

```go
// 创建时间范围条件
startTime := time.Now().AddDate(0, -1, 0) // 一个月前
endTime := time.Now()
timeCondition := query.NewTimeRangeCondition("created_at", &startTime, &endTime)

// 创建搜索条件
searchCondition := query.NewSearchCondition("张三", []string{"name", "nickname", "email"})

// 组合条件
builder := factory.NewQuery().
    Where(timeCondition).
    Where(searchCondition).
    WhereSimple("status", query.OpEq, "active")

// 创建分页请求
pageReq := query.NewPageRequest(1, 20)
pageReq.OrderBy = "created_at"
pageReq.Order = "DESC"

// 执行查询
// ...
```

### 从HTTP请求参数构建查询条件

```go
// 假设有一个HTTP请求处理函数
func ListUsers(c *fiber.Ctx) error {
    // 解析分页参数
    page, _ := strconv.Atoi(c.Query("page", "1"))
    pageSize, _ := strconv.Atoi(c.Query("pageSize", "10"))
    pageReq := query.NewPageRequest(page, pageSize)
    pageReq.OrderBy = c.Query("orderBy", "created_at")
    pageReq.Order = c.Query("order", "DESC")
    
    // 创建查询构建器
    builder := factory.NewQuery()
    
    // 从请求参数构建查询条件
    if status := c.Query("status"); status != "" {
        builder.WhereSimple("status", query.OpEq, status)
    }
    
    if name := c.Query("name"); name != "" {
        builder.WhereSimple("name", query.OpContains, name)
    }
    
    if minAge, err := strconv.Atoi(c.Query("minAge", "0")); err == nil && minAge > 0 {
        builder.WhereSimple("age", query.OpGte, minAge)
    }
    
    // 执行查询
    var users []User
    pageRes, err := query.Paginate(ctx, builder, provider, pageReq, &users)
    if err != nil {
        return err
    }
    
    return c.JSON(pageRes)
}
```

## 事务示例

```go
provider := query.NewMySQLProvider[User](db)

err := provider.Transaction(ctx, func(ctx context.Context) error {
    // 在事务中执行操作
    user := User{Name: "张三", Age: 30}
    
    // 插入记录
    if err := provider.Insert(ctx, &user); err != nil {
        return err
    }
    
    // 更新记录
    builder := factory.NewQuery().WhereSimple("id", query.OpEq, user.ID)
    updates := map[string]interface{}{"status": "active"}
    if err := provider.Update(ctx, builder.Build(), updates); err != nil {
        return err
    }
    
    return nil
})

if err != nil {
    // 处理事务错误
}
```

## 注意事项

1. MongoDB的聚合查询（如GROUP BY、HAVING等）需要使用聚合管道，当前实现做了简化处理
2. 对于复杂的原生查询，可以使用WhereRaw方法传入原始查询条件
3. 查询构建器和数据提供者是解耦的，可以单独使用
4. 所有查询方法都支持上下文（context），可以用于超时控制和取消操作
5. 分页请求（PageRequest）只负责分页和排序，不再包含过滤条件，过滤条件应通过查询构建器（QueryBuilder）设置 