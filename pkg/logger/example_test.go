package logger

import (
	"errors"
	"time"
)

// ExampleLogger_new_api 展示新的 logger API 使用方式
func ExampleLogger_new_api() {
	// 使用新的 Field 构造函数，不再需要导入 zap
	Info("User login successful",
		String("user_id", "12345"),
		String("username", "john_doe"),
		Int("login_attempts", 1),
		Duration("response_time", 250*time.Millisecond),
		Bool("is_admin", false),
	)

	// 错误日志
	ErrorLog("Database connection failed",
		ErrorField(errors.New("connection timeout")),
		String("database", "postgres"),
		Int("retry_count", 3),
	)

	// 使用 With 创建子 logger
	userLogger := With(
		String("service", "user_service"),
		String("version", "v1.2.3"),
	)

	userLogger.Info("Processing user data",
		String("operation", "update_profile"),
		Int64("user_id", 12345),
	)

	// 调试信息
	Debug("Cache operation",
		String("cache_key", "user:12345"),
		String("operation", "get"),
		Bool("cache_hit", true),
		Duration("latency", 5*time.Millisecond),
	)

	// 警告信息
	Warn("Rate limit approaching",
		String("client_ip", "192.168.1.100"),
		Int("current_requests", 95),
		Int("limit", 100),
		Float64("usage_percent", 95.0),
	)

	// 二进制数据
	Debug("Received data packet",
		Binary("payload", []byte{0x01, 0x02, 0x03}),
		Int("packet_size", 3),
	)

	// 任意类型数据
	Info("Custom event",
		Any("metadata", map[string]interface{}{
			"event_type": "user_action",
			"timestamp":  time.Now(),
			"properties": []string{"click", "button", "submit"},
		}),
	)
}

// ExampleLogger_migration 展示如何从现有的 zap.Field 迁移
func ExampleLogger_migration() {
	// 如果你有现有的 zap.Field，可以使用 FromZapField 进行迁移
	// 注意：这只是用于过渡期间，推荐最终完全迁移到新的 API

	// 模拟现有的 zap.Field（实际使用中这些可能来自其他函数）
	// zapFields := []zap.Field{
	//     zap.String("service", "api"),
	//     zap.Int("port", 8080),
	//     zap.Error(errors.New("some error")),
	// }

	// 转换为新的 Field
	// fields := FromZapFields(zapFields)
	// Info("Migrated log", fields...)

	// 推荐的新方式
	Info("Service started",
		String("service", "api"),
		Int("port", 8080),
		String("environment", "production"),
	)
}

// ExampleLogger_performance 展示性能优化的使用方式
func ExampleLogger_performance() {
	// 对于高频日志，可以预先创建字段
	serviceField := String("service", "high_frequency_service")
	versionField := String("version", "v2.0.0")

	// 在循环中复用字段
	for i := 0; i < 1000; i++ {
		Debug("Processing item",
			serviceField,
			versionField,
			Int("item_id", i),
			Time("processed_at", time.Now()),
		)
	}

	// 使用 With 创建带有公共字段的 logger
	requestLogger := With(
		String("request_id", "req-123456"),
		String("user_agent", "MyApp/1.0"),
		String("remote_addr", "192.168.1.100"),
	)

	// 请求处理过程中的日志都会自动包含上述字段
	requestLogger.Info("Request started")
	requestLogger.Debug("Validating input")
	requestLogger.Info("Request completed", Duration("duration", 150*time.Millisecond))
}
