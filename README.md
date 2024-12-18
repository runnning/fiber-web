# Fiber Web Framework Scaffold

A production-ready Go web application scaffold using Fiber framework with clean architecture.

## Features

- Clean Architecture
- Fiber Web Framework
- GORM with MySQL
- Redis Integration
  - Caching support
  - Key-value and hash operations
  - Connection pooling
- Distributed Lock
  - Redis-based implementation
  - Auto-refresh support
  - Safe lock release
  - Extensible interface
- Authentication & Authorization
  - JWT-based authentication
  - Role-based access control (RBAC)
  - Token refresh mechanism
  - Middleware protection
- NSQ Message Queue
  - Producer/Consumer pattern
  - Delayed message support
  - Multiple topics and channels
- Cron Task Scheduler
  - Task timeout control
  - Concurrent execution
  - LRU-based task management
  - Graceful shutdown support
- Zap Logger Integration
  - Environment-based configuration
  - File and console output
  - Log rotation by date
  - Structured logging
- Generic Utility Packages
  - Slice operations (Map, Filter, Reduce, etc.)
  - Map utilities (Keys, Values, Merge, etc.)
  - Enhanced error handling with stack traces
  - Functional programming tools (Pipe, Compose, Either, Option)
  - Concurrent programming utilities (Worker Pool, SafeMap, etc.)
  - Time utilities
    - Date/Time formatting and parsing
    - Time boundary calculations (day/week/month)
    - Age and duration calculations
    - Relative time descriptions
    - Working days handling
- Configuration Management (Viper)
- Error Handling
- Connection Pool Management
- Middleware Support
- RESTful API Design

## Project Structure

```
.
├── cmd/                    # Application entry points
│   ├── api/               # API server
│   │   └── main.go
│   └── config/           # Configuration files
│       ├── config.local.yaml
│       └── config.docker.yaml
├── deploy/                # Deployment configurations
│   └── Dockerfile
├── internal/              # Private application code
│   ├── entity/           # Enterprise business rules (entities)
│   ├── usecase/          # Application business rules
│   ├── repository/       # Interface adapters (database implementations)
│   ├── endpoint/         # Interface adapters (HTTP handlers)
│   ├── middleware/       # Custom middleware
│   └── initialize/       # Application initialization
├── pkg/                   # Public libraries
│   ├── config/           # Configuration management
│   ├── cron/             # Task scheduler
│   ├── database/         # Database connections
│   ├── logger/           # Logging utilities
│   ├── redis/            # Redis client
│   ├── lock/             # Distributed lock
│   ├── auth/             # Authentication & Authorization
│   ├── queue/            # Message queue
│   ├── server/           # Fiber server
│   ├── color/            # Color utilities
│   ├── http_cli/         # Http client
│   ├── validator/        # Validator
│   ├── ctx/              # Context utilities
│   ├── query/            # Query utilities
│   └── utils/            # Utility packages
│       ├── concurrent/   # Concurrent programming utilities
│       ├── errorx/       # Enhanced error handling
│       ├── fp/           # Functional programming tools
│       ├── maps/         # Map operations
│       ├── slice/        # Slice operations
│       └── time_util/     # Time utilities
│       └── file_util/     # File utilities
├── tools/                 # Development tools
│   ├── generator/        # Code generator
│   │   └── templates/    # Code templates
│   │       ├── endpoint.go
│   │       ├── repository.go
│   │       └── usecase.go
│   │       └── entity.go
│   └── main.go
└── logs/                  # Application logs
```

## Prerequisites

- Go 1.21 or higher
- MySQL 5.7 or higher
- Redis 6.0 or higher
- NSQ v1.2 or higher

## Quick Start

1. Clone the repository:
```bash
git clone <repository-url>
cd fiber-web
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set up services:

```bash
# Start Redis
docker run -d --name redis -p 6379:6379 redis

# Start NSQ
docker run -d --name nsqlookupd -p 4160:4160 -p 4161:4161 nsqio/nsq /nsqlookupd
docker run -d --name nsqd -p 4150:4150 -p 4151:4151 \
    nsqio/nsq /nsqd \
    --broadcast-address=localhost \
    --lookupd-tcp-address=localhost:4160

# Create MySQL database
mysql -u root -p
CREATE DATABASE fiber_web CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

4. Configure the application:

```bash
cp config.yaml.example config.yaml
```

Edit `config.yaml`:
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

5. Run the application:
```bash
go run cmd/api/main.go
```

## Code Generation Tool

The project includes a code generation tool that quickly generates CRUD code structure based on clean architecture principles.

### Usage

1. Generate new entity and related code:

```bash
# Basic usage
go run tools/main.go -name EntityName -module ModuleName

# Examples:
# Generate User entity in auth module
go run tools/main.go -name User -module auth

# Generate Product entity in shop module
go run tools/main.go -name Product -module shop

# Generate Order entity in order module
go run tools/main.go -name Order -module order
```

### Generated Structure

For each module, the tool will generate the following structure:
```
module_name/
├── entity/                 # Domain entities
│   └── entity_name.go
├── repository/            # Data access interfaces
│   └── entity_name_repository.go
├── usecase/              # Business logic
│   └── entity_name_usecase.go
└── endpoint/             # HTTP handlers
    └── entity_name_endpoint.go
```

### Generated Code Features

- Complete CRUD operations
- Clean Architecture structure
- Interface-based design
- Standard Go project layout
- Ready-to-use HTTP endpoints

### Important Notes

- Generated code uses the module name in import paths
- Files are generated with .tpl extension
- Each module maintains its own clean architecture structure
- Customize the generated code according to your business needs

## API Documentation

### Authentication

```bash
# Register a new user
curl -X POST http://localhost:3000/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"secret","email":"john@example.com"}'

# Login
curl -X POST http://localhost:3000/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"secret"}'

# Refresh token
curl -X POST http://localhost:3000/api/v1/refresh \
  -H "Authorization: Bearer <refresh_token>"
```

### Protected Routes

```bash
# Get user profile
curl -X GET http://localhost:3000/api/v1/users/me \
  -H "Authorization: Bearer <access_token>"

# Update user
curl -X PUT http://localhost:3000/api/v1/users/me \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"email":"new.email@example.com"}'
```

## Component Usage

### Redis Cache

```go
import "fiber_web/pkg/redis"

// Store data with expiration
err := redisClient.Set(ctx, "user:123", userJSON, time.Hour)

// Retrieve data
val, err := redisClient.Get(ctx, "user:123")

// Hash operations
err := redisClient.HSet(ctx, "user:123", "name", "John", "age", "30")
name, err := redisClient.HGet(ctx, "user:123", "name")
```

### Distributed Lock

```go
import "fiber_web/pkg/lock"

// Create a Redis lock
lock, err := lock.NewRedisLock(lock.Options{
    Key:         "my-resource",
    Expiration:  5 * time.Second,
    RedisClient: redisClient,
})

// Try to acquire the lock
acquired, err := lock.TryLock(ctx)
if acquired {
    // Enable auto-refresh
    cancel, err := lock.AutoRefresh(ctx, time.Second)
    defer cancel()
    
    // Do work here...
    
    // Release the lock
    err = lock.Unlock(ctx)
}
```

### JWT Authentication

```go
import "fiber_web/pkg/auth"

// Initialize JWT manager
jwtManager := auth.NewJWTManager(cfg)

// Generate token pair
tokenPair, err := jwtManager.GenerateTokenPair(user.ID, user.Username, user.Role)

// Protect routes
app.Use("/api/v1", auth.Protected(jwtManager))

// Refresh token
newTokenPair, err := jwtManager.RefreshToken(refreshToken)
```

### Casbin Authorization

```go
import "fiber_web/pkg/auth"

// Initialize Casbin
enforcer, err := auth.InitCasbin(db, cfg)

// Protect routes
app.Use("/api/v1", auth.Authorize())

// Manage permissions
auth.AddPolicy("admin", "/api/v1/users", "POST")
auth.AddRoleForUser("user123", "admin")
allowed := auth.HasPermission("user123", "/api/v1/users", "GET")

// Advanced policies
auth.AddPolicy("manager", "/api/v1/reports/*", "GET")  // 通配符匹配
auth.AddPolicy("editor", "/api/v1/posts/:id", "*")     // 支持所有方法

// Role inheritance
auth.AddRoleForUser("admin", "manager")    // admin 继承 manager 的权限
auth.AddRoleForUser("manager", "editor")   // manager 继承 editor 的权限

// Check permissions
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

### NSQ Message Queue

```go
import "fiber_web/pkg/queue"

// Producer
producer, err := queue.NewProducer(cfg)
err = producer.Publish("user_events", messageBytes)
err = producer.PublishDeferred("reminders", time.Hour, messageBytes)

// Consumer
handler := nsq.HandlerFunc(func(message *nsq.Message) error {
    log.Printf("Got message: %s", message.Body)
    return nil
})
consumer, err := queue.NewConsumer("user_events", "email_service", handler, cfg)
defer consumer.Stop()
```

### Cron Task Scheduler

```go
import "fiber_web/pkg/cron"

// Initialize scheduler
scheduler := cron.NewScheduler(logger)

// Add a task with timeout
err := scheduler.AddTask(
    "daily-backup",           // Task name
    "0 0 3 * * *",           // Run at 3 AM daily
    func() error {           // Task function
        // Perform backup operation
        return nil
    },
    30*time.Minute,          // Timeout duration
)

// Start the scheduler
scheduler.Start()
defer scheduler.Stop()

// Get task information
task, err := scheduler.GetTask("daily-backup")

// List all tasks
tasks := scheduler.ListTasks()

// Remove a task
err = scheduler.RemoveTask("daily-backup")
```

### Generic Utilities

#### Slice Operations
```go
import "fiber_web/pkg/utils/slice"

// Map: Transform each element
numbers := []int{1, 2, 3, 4, 5}
doubled := slice.Map(numbers, func(x int) int { return x * 2 })
// Result: [2, 4, 6, 8, 10]

// Filter: Keep elements that satisfy a predicate
evens := slice.Filter(numbers, func(x int) bool { return x%2 == 0 })
// Result: [2, 4]

// Reduce: Accumulate elements
sum := slice.Reduce(numbers, 0, func(acc, x int) int { return acc + x })
// Result: 15

// Sort: Sort elements in ascending order
sorted := slice.Sort([]int{3, 1, 4, 1, 5})
// Result: [1, 1, 3, 4, 5]
```

#### Map Operations
```go
import "fiber_web/pkg/utils/maps"

users := map[string]int{
    "alice": 20,
    "bob":   25,
    "carol": 30,
}

// Get all keys
names := maps.Keys(users)
// Result: ["alice", "bob", "carol"]

// Filter users by age
adults := maps.Filter(users, func(name string, age int) bool {
    return age >= 21
})
// Result: {"bob": 25, "carol": 30}

// Transform values
ageNextYear := maps.MapValues(users, func(age int) int {
    return age + 1
})
// Result: {"alice": 21, "bob": 26, "carol": 31}
```

#### Error Handling
```go
import "fiber_web/pkg/utils/errorx"

// Create custom error with context
err := errorx.New("connection failed").
    WithCode(500).
    WithOperation("DatabaseConnect").
    WithContext("host", "localhost")

// Safe function execution
result, err := errorx.Try(func() string {
    // potentially panicking code
    return "success"
})

// Must will panic if error occurs
value := errorx.Must(strconv.Atoi("123"))
```

#### Functional Programming
```go
import "fiber_web/pkg/utils/fp"

// Function composition
add1 := func(x int) int { return x + 1 }
multiply2 := func(x int) int { return x * 2 }
pipeline := fp.Compose(add1, multiply2)
result := pipeline(5) // ((5 + 1) * 2) = 12

// Option type for nullable values
user := fp.Some("John")
if user.IsSome() {
    name := user.Unwrap()
}

// Either type for error handling
result := fp.Right[string, int](42)
result.Match(
    func(err string) { fmt.Println("Error:", err) },
    func(val int) { fmt.Println("Success:", val) },
)
```

#### Concurrent Programming
```go
import "fiber_web/pkg/utils/concurrent"

// Worker pool
pool := concurrent.NewPool[int](5, 10)
pool.Start()
pool.Submit(func() int { return 42 })
for result := range pool.Results() {
    fmt.Println(result)
}

// Thread-safe map
safeMap := concurrent.NewSafeMap[string, int]()
safeMap.Set("counter", 1)
value, exists := safeMap.Get("counter")

// Parallel execution
results := concurrent.Parallel(
    func() int { return 1 },
    func() int { return 2 },
    func() int { return 3 },
)

// Debounce function calls
debouncedSave := concurrent.Debounce(func(data string) {
    // save to database
}, time.Second)
```

#### Time Utilities
```go
import "fiber_web/pkg/utils/timeutil"

// Format dates and times
now := time.Now()
dateStr := timeutil.FormatDate(now)         // "2024-03-14"
timeStr := timeutil.FormatDateTime(now)     // "2024-03-14 15:04:05"

// Get time boundaries
dayStart := timeutil.StartOfDay(now)        // 2024-03-14 00:00:00
monthStart := timeutil.StartOfMonth(now)    // 2024-03-01 00:00:00
weekStart := timeutil.StartOfWeek(now)      // Monday 00:00:00

// Calculate durations and ages
birthDate, _ := timeutil.ParseDate("1990-01-01")
age := timeutil.Age(birthDate)              // 34

duration := 25 * time.Hour
formatted := timeutil.FormatDuration(duration) // "1天1小时0分钟"

// Get relative time descriptions
posted := time.Now().Add(-2 * time.Hour)
relative := timeutil.RelativeTime(posted)    // "2小时前"

// Working days
nextWorkDay := timeutil.AddWorkDays(now, 1)  // Skip weekends
isWeekend := timeutil.IsWeekend(now)        // Check if weekend
```

## Development

### Adding New Features

1. Define domain entities and interfaces in `internal/domain`
2. Implement use cases in `internal/usecase`
3. Add repository implementation in `internal/repository`
4. Create HTTP handlers in `internal/delivery`
5. Use utility packages from `pkg/utils` to simplify implementation

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

### Logging

The application uses structured logging with different levels:

```go
logger.Debug("Processing request", zap.Any("payload", payload))
logger.Info("User action", zap.String("user_id", "123"))
logger.Error("Operation failed", zap.Error(err))
```

## Production Deployment

1. Set appropriate environment variables
2. Use secure values for all credentials
3. Enable HTTPS
4. Set up monitoring and alerting
5. Configure log rotation
6. Use connection pooling
7. Enable rate limiting

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License.

## Configuration

The project uses different configuration files for different environments:

- `config.local.yaml`: Local development environment (default)
- `config.docker.yaml`: Docker container environment

You can specify which configuration to use by setting the `CONFIG_NAME` environment variable:

```bash
# Local development
CONFIG_NAME=config.local go run cmd/api/main.go

# Docker environment
CONFIG_NAME=config.docker go run cmd/api/main.go
```

In Docker, the configuration is automatically set to use the Docker environment settings.
