# Fiber Web 框架脚手架

[English](README-EN.md) | 简体中文

一个基于 Fiber 框架的生产级 Go web 应用脚手架，采用清晰架构设计。

## 特性

- 清晰架构
- Fiber Web 框架
- 使用 MySQL 的 GORM
- Redis 集成
  - 缓存支持
  - 键值对和哈希操作
  - 连接池
- 分布式锁
  - 基于 Redis 的实现
  - 自动刷新支持
  - 安全锁释放
  - 可扩展接口
- 认证与授权
  - 基于 JWT 的认证
  - 基于角色的访问控制 (RBAC)
  - 令牌刷新机制
  - 中间件保护
- NSQ 消息队列
  - 生产者/消费者模式
  - 延迟消息支持
  - 多主题和通道
- Cron 任务调度器
  - 任务超时控制
  - 并发执行
  - 基于 LRU 的任务管理
  - 优雅关闭支持
- Zap 日志集成
  - 基于环境的配置
  - 文件和控制台输出
  - 按日期轮转日志
  - 结构化日志
- 通用工具包
  - 切片操作 (Map, Filter, Reduce 等)
  - Map 工具 (Keys, Values, Merge 等)
  - 增强的错误处理与堆栈跟踪
  - 函数式编程工具 (Pipe, Compose, Either, Option)
  - 并发编程工具 (工作池, 安全Map等)
  - 时间工具
    - 日期/时间格式化和解析
    - 时间边界计算 (天/周/月)
    - 年龄和时长计算
    - 相对时间描述
    - 工作日处理
- 配置管理 (Viper)
- 错误处理
- 连接池管理
- 中间件支持
- RESTful API 设���

## 项目结构

```
.
├── apps/                   # 应用目录
│   ├── admin/             # 管理服务
│   │   ├── cmd/          # 入口点
│   │   │   ├── api/      # API 服务入口
│   │   │   │   └── main.go
│   │   │   └── config/   # 配置文件
│   │   └── internal/     # 内部代码
│   │       ├── endpoint/ # HTTP 处理器
│   │       ├── bootstrap/ # 引导程序
│   │       ├── middleware/ # 中间件
│   │       ├── entity/   # 领域实体
│   │       ├── initialize/ # 应用初始化
│   │       ├── repository/ # 数据仓库
│   │       └── usecase/   # 业务逻辑
├── deploy/               # 部署配置
│   ├── Dockerfile
│   └── docker-compose.yml
├── pkg/                  # 公共库
│   ├── auth/            # 认证与授权
│   ├── config/          # 配置管理
│   ├── database/        # 数据库连接
│   ├── logger/          # 日志工具
│   ├── redis/           # Redis 客户端
│   ├── lock/            # 分布式锁
│   ├── queue/           # 消息队列
│   ├── server/          # HTTP 服务器
│   ├── validator/       # 验证器
│   ��── color/           # 颜色工具
│   ├── ctx/             # 上下文工具
│   ├── query/           # 查询工具
│   ├── response/        # 响应工具
│   ├── server/          # 服务器工具
│   ├── security/        # 安全工具
│   ├── logger/          # 日志工具
│   ├── cron/            # 定时任务工具
│   ├── http_cli/        # HTTP 客户端工具
│   └── utils/           # 工具包
│       ├── concurrent/  # 并发工具
│       ├── errorx/      # 错误处理
│       ├── maps/        # Map 操作
│       ├── slice/       # 切片操作
│       └── time_util/   # 时间工具
│       └── file_util/   # 文件工具
│       └── fp/          # 函数式编程
└── tools/               # 开发工具
    └── generator/       # 代码生成器
```

## 前置条件

- Go 1.21 或更高版本
- MySQL 5.7 或更高版本
- Redis 6.0 或更高版本
- NSQ v1.2 或更高版本

## 快速开始

1. 克隆仓库:
```bash
git clone <仓库地址>
cd fiber-web
```

2. 安装依赖:
```bash
go mod tidy
```

3. 设置服务:

```bash
# 启动 Redis
docker run -d --name redis -p 6379:6379 redis

# 启动 NSQ
docker run -d --name nsqlookupd -p 4160:4160 -p 4161:4161 nsqio/nsq /nsqlookupd
docker run -d --name nsqd -p 4150:4150 -p 4151:4151 \
    nsqio/nsq /nsqd \
    --broadcast-address=localhost \
    --lookupd-tcp-address=localhost:4160

# 创建 MySQL 数据库
mysql -u root -p
CREATE DATABASE fiber_web CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

4. 配置应用:

```bash
cp config.yaml.example config.yaml
```

编辑 `config.yaml`:
```yaml
server:
  address: ":3000"
  port: 3000

database:
  host: "localhost"
  port: 3306
  user: "root"
  password: "root"
  dbname: "fiber_web"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

nsq:
  nsqd:
    host: "localhost"
    port: 4150
  lookupd:
    host: "localhost"
    port: 4161

app:
  env: "development"
  name: "fiber-web"

jwt:
  secret_key: "your-secret-key-here"
  access_token_expiry: "15m"
  refresh_token_expiry: "168h"
```

5. 运行应用:
```bash
go run cmd/api/main.go
```

## 代码生成工具

项目包含一个代码生成工具，可以基于清晰架构原则快速生成 CRUD 代码结构。

### 使用方法

1. 生成新实体和相关代码:

```bash
# 基本用法
go run tools/main.go -name 实体名 -module 模块名

# 示例:
# 在 auth 模块中生成 User 实体
go run tools/main.go -name User -module auth

# 在 shop 模块中生成 Product 实体
go run tools/main.go -name Product -module shop

# 在 order 模块中生成 Order 实体
go run tools/main.go -name Order -module order
```

### 生成的结构

对于每个模块，工具将生成以下结构:
```
module_name/
├── entity/                 # 领域实体
│   └── entity_name.go
├── repository/            # 数据访问接口
│   └── entity_name_repository.go
├── usecase/              # 业务逻辑
│   └── entity_name_usecase.go
└── endpoint/             # HTTP 处理器
    └── entity_name_endpoint.go
```

### 生成的代码特性

- 完整的 CRUD 操作
- 清晰架构结构
- 基于接口的设计
- 标准 Go 项目布局
- 即用型 HTTP 端点

### 重要说明

- 生成的代码使用模块名作为导入路径
- 文件以 .tpl 扩展名生成
- 每个模块维护自己的清晰架构结构
- 根据业务需求自定义生成的代码

## API 文档

### 认证

```bash
# 注册新用户
curl -X POST http://localhost:3000/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"secret","email":"john@example.com"}'

# 登录
curl -X POST http://localhost:3000/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"secret"}'

# 刷新令牌
curl -X POST http://localhost:3000/api/v1/refresh \
  -H "Authorization: Bearer <refresh_token>"
```

### 受保护的路由

```bash
# 获取用户资料
curl -X GET http://localhost:3000/api/v1/users/me \
  -H "Authorization: Bearer <access_token>"

# 更新用户
curl -X PUT http://localhost:3000/api/v1/users/me \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"email":"new.email@example.com"}'
```

## 组件使用

### Redis 缓存

```go
import "fiber_web/pkg/redis"

// 存储带过期时间的数据
err := redisClient.Set(ctx, "user:123", userJSON, time.Hour)

// 获取数据
val, err := redisClient.Get(ctx, "user:123")

// 哈希操作
err := redisClient.HSet(ctx, "user:123", "name", "John", "age", "30")
name, err := redisClient.HGet(ctx, "user:123", "name")
```

### 分布式锁

```go
import "fiber_web/pkg/lock"

// 创建 Redis 锁
lock, err := lock.NewRedisLock(lock.Options{
    Key:         "my-resource",
    Expiration:  5 * time.Second,
    RedisClient: redisClient,
})

// 尝试获取锁
acquired, err := lock.TryLock(ctx)
if acquired {
    // 启用自动刷新
    cancel, err := lock.AutoRefresh(ctx, time.Second)
    defer cancel()
    
    // 在这���执行工作...
    
    // 释放锁
    err = lock.Unlock(ctx)
}
```

### JWT 认证

```go
import "fiber_web/pkg/auth"

// 初始化 JWT 管理器
jwtManager := auth.NewJWTManager(cfg)

// 生成令牌对
tokenPair, err := jwtManager.GenerateTokenPair(user.ID, user.Username, user.Role)

// 保护路由
app.Use("/api/v1", auth.Protected(jwtManager))

// 刷新令牌
newTokenPair, err := jwtManager.RefreshToken(refreshToken)
```

### Casbin 授权

```go
import "fiber_web/pkg/auth"

// 初始化 Casbin
enforcer, err := auth.InitCasbin(db, cfg)

// 保护路由
app.Use("/api/v1", auth.Authorize())

// 管理权限
auth.AddPolicy("admin", "/api/v1/users", "POST")
auth.AddRoleForUser("user123", "admin")
allowed := auth.HasPermission("user123", "/api/v1/users", "GET")

// 高级策略
auth.AddPolicy("manager", "/api/v1/reports/*", "GET")  // 通配符匹配
auth.AddPolicy("editor", "/api/v1/posts/:id", "*")     // 支持所有方法

// 角色继承
auth.AddRoleForUser("admin", "manager")    // admin 继承 manager 的权限
auth.AddRoleForUser("manager", "editor")   // manager 继承 editor 的权限

// 检查权限
roles, _ := auth.GetRolesForUser("user123")
users, _ := auth.GetUsersForRole("admin")
```

RBAC 模型配置:
```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && (r.act == p.act || p.act == "*")
```

### NSQ 消息队列

```go
import "fiber_web/pkg/queue"

// 生产者
producer, err := queue.NewProducer(cfg)
err = producer.Publish("user_events", messageBytes)
err = producer.PublishDeferred("reminders", time.Hour, messageBytes)

// 消费者
handler := nsq.HandlerFunc(func(message *nsq.Message) error {
    log.Printf("收到消息: %s", message.Body)
    return nil
})
consumer, err := queue.NewConsumer("user_events", "email_service", handler, cfg)
defer consumer.Stop()
```

### Cron 任务调度器

```go
import "fiber_web/pkg/cron"

// 初始化调度器
scheduler := cron.NewScheduler(logger)

// 添加带超时的任务
err := scheduler.AddTask(
    "daily-backup",           // 任务名称
    "0 0 3 * * *",           // 每天凌晨3点运行
    func() error {           // 任务函数
        // 执行备份操作
        return nil
    },
    30*time.Minute,          // 超时时间
)

// 启动调度器
scheduler.Start()
defer scheduler.Stop()

// 获取任务信息
task, err := scheduler.GetTask("daily-backup")

// 列出所有任务
tasks := scheduler.ListTasks()

// 移除任务
err = scheduler.RemoveTask("daily-backup")
```

### 通用工具

#### 切片操作
```go
import "fiber_web/pkg/utils/slice"

// Map: 转换每个元素
numbers := []int{1, 2, 3, 4, 5}
doubled := slice.Map(numbers, func(x int) int { return x * 2 })
// 结果: [2, 4, 6, 8, 10]

// Filter: 保留满足条件的元素
evens := slice.Filter(numbers, func(x int) bool { return x%2 == 0 })
// 结果: [2, 4]

// Reduce: 累积元素
sum := slice.Reduce(numbers, 0, func(acc, x int) int { return acc + x })
// 结果: 15

// Sort: 升序排序
sorted := slice.Sort([]int{3, 1, 4, 1, 5})
// 结果: [1, 1, 3, 4, 5]
```

#### Map 操作
```go
import "fiber_web/pkg/utils/maps"

users := map[string]int{
    "alice": 20,
    "bob":   25,
    "carol": 30,
}

// 获取所有键
names := maps.Keys(users)
// 结果: ["alice", "bob", "carol"]

// 按年龄过滤用户
adults := maps.Filter(users, func(name string, age int) bool {
    return age >= 21
})
// 结果: {"bob": 25, "carol": 30}

// 转换值
ageNextYear := maps.MapValues(users, func(age int) int {
    return age + 1
})
// 结果: {"alice": 21, "bob": 26, "carol": 31}
```

#### 错误处理
```go
import "fiber_web/pkg/utils/errorx"

// 创建带上下文的自定义错误
err := errorx.New("连接失败").
    WithCode(500).
    WithOperation("DatabaseConnect").
    WithContext("host", "localhost")

// 安全函数执行
result, err := errorx.Try(func() string {
    // 可能会发生panic的代码
    return "success"
})

// Must 在发生错误时会panic
value := errorx.Must(strconv.Atoi("123"))
```

#### 函数式编程
```go
import "fiber_web/pkg/utils/fp"

// 函数组合
add1 := func(x int) int { return x + 1 }
multiply2 := func(x int) int { return x * 2 }
pipeline := fp.Compose(add1, multiply2)
result := pipeline(5) // ((5 + 1) * 2) = 12

// Option ��型用于可空值
user := fp.Some("John")
if user.IsSome() {
    name := user.Unwrap()
}

// Either 类型用于错误处理
result := fp.Right[string, int](42)
result.Match(
    func(err string) { fmt.Println("错误:", err) },
    func(val int) { fmt.Println("成功:", val) },
)
```

#### 并发编程
```go
import "fiber_web/pkg/utils/concurrent"

// 工作池
pool := concurrent.NewPool[int](5, 10)
pool.Start()
pool.Submit(func() int { return 42 })
for result := range pool.Results() {
    fmt.Println(result)
}

// 线程安全的map
safeMap := concurrent.NewSafeMap[string, int]()
safeMap.Set("counter", 1)
value, exists := safeMap.Get("counter")

// 并行执行
results := concurrent.Parallel(
    func() int { return 1 },
    func() int { return 2 },
    func() int { return 3 },
)

// 防抖函数调用
debouncedSave := concurrent.Debounce(func(data string) {
    // 保存到数据库
}, time.Second)
```

#### 时间工具
```go
import "fiber_web/pkg/utils/timeutil"

// 格式化日期和时间
now := time.Now()
dateStr := timeutil.FormatDate(now)         // "2024-03-14"
timeStr := timeutil.FormatDateTime(now)     // "2024-03-14 15:04:05"

// 获取时间边界
dayStart := timeutil.StartOfDay(now)        // 2024-03-14 00:00:00
monthStart := timeutil.StartOfMonth(now)    // 2024-03-01 00:00:00
weekStart := timeutil.StartOfWeek(now)      // 周一 00:00:00

// 计算时长和年龄
birthDate, _ := timeutil.ParseDate("1990-01-01")
age := timeutil.Age(birthDate)              // 34

duration := 25 * time.Hour
formatted := timeutil.FormatDuration(duration) // "1天1小时0分钟"

// 获取相对时间描述
posted := time.Now().Add(-2 * time.Hour)
relative := timeutil.RelativeTime(posted)    // "2小时前"

// 工作日
nextWorkDay := timeutil.AddWorkDays(now, 1)  // 跳过周末
isWeekend := timeutil.IsWeekend(now)        // 检查是否为周末
```

## 开发

### 添加新功能

1. 在 `internal/domain` 中定义领域实体和接口
2. 在 `internal/usecase` 中实现用例
3. 在 `internal/repository` 中添加仓库实现
4. 在 `internal/delivery` 中创建 HTTP 处理器
5. 使用 `pkg/utils` 中的工具包简化实现

### 测试

```bash
# 运行所有测试
go test ./...

# 运行带覆盖率的测试
go test -cover ./...
```

### 日志记录

应用使用不同级别的结构化日志：

```go
logger.Debug("处理请求", zap.Any("payload", payload))
logger.Info("用户操作", zap.String("user_id", "123"))
logger.Error("操作失败", zap.Error(err))
```

## 生产部署

1. 设置适当的环境变量
2. 使用安全的凭证值
3. 启用 HTTPS
4. 设置监控和告警
5. 配置日志轮转
6. 使用连接池
7. 启用速率限制

## 贡献

1. Fork 仓库
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建新的 Pull Request

## 许可证

本项目采用 MIT 许可证。

## 配置

项目使用不同的配置文件用于不同环境：

- `config.local.yaml`: 本地开发环境（默认）
- `config.docker.yaml`: Docker 容器环境

你可以通过设置 `CONFIG_NAME` 环境变量来指定使用哪个配置：

```bash
# 本地开发
CONFIG_NAME=config.local go run cmd/api/main.go

# Docker 环境
CONFIG_NAME=config.docker go run cmd/api/main.go
```

在 Docker 中，配置会自动设置为使用 Docker 环境设置。
