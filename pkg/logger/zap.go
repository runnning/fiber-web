package logger

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"fiber_web/pkg/config"
	"fiber_web/pkg/database"
	"fiber_web/pkg/utils/concurrent"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultMaxSize    = 100 // 默认最大文件大小（MB）
	defaultMaxBackups = 3   // 默认最大备份数
	defaultMaxAge     = 28  // 默认最大保存天数

	// 异步日志配置
	defaultAsyncWorkers = 4    // 默认异步工作者数量
	defaultAsyncBuffer  = 1000 // 默认异步缓冲区大小
)

// FieldType 字段类型枚举
type FieldType int

const (
	StringFieldType FieldType = iota
	IntFieldType
	Int64FieldType
	Float64FieldType
	BoolFieldType
	ErrorFieldType
	TimeFieldType
	DurationFieldType
	BinaryFieldType
	ByteStringFieldType
	AnyFieldType
	SkipFieldType
)

// Field 代表一个日志字段的抽象接口
type Field interface {
	Key() string
	Value() interface{}
	Type() FieldType
	toZapField() zap.Field
}

// logField 实现 Field 接口
type logField struct {
	key       string
	value     interface{}
	fieldType FieldType
}

func (f *logField) Key() string {
	return f.key
}

func (f *logField) Value() interface{} {
	return f.value
}

func (f *logField) Type() FieldType {
	return f.fieldType
}

// toZapField 将自定义 Field 转换为 zap.Field
func (f *logField) toZapField() zap.Field {
	switch f.fieldType {
	case StringFieldType:
		return zap.String(f.key, f.value.(string))
	case IntFieldType:
		return zap.Int(f.key, f.value.(int))
	case Int64FieldType:
		return zap.Int64(f.key, f.value.(int64))
	case Float64FieldType:
		return zap.Float64(f.key, f.value.(float64))
	case BoolFieldType:
		return zap.Bool(f.key, f.value.(bool))
	case ErrorFieldType:
		return zap.Error(f.value.(error))
	case TimeFieldType:
		return zap.Time(f.key, f.value.(time.Time))
	case DurationFieldType:
		return zap.Duration(f.key, f.value.(time.Duration))
	case BinaryFieldType:
		return zap.Binary(f.key, f.value.([]byte))
	case ByteStringFieldType:
		return zap.ByteString(f.key, f.value.([]byte))
	case AnyFieldType:
		return zap.Any(f.key, f.value)
	case SkipFieldType:
		return zap.Skip()
	default:
		return zap.Any(f.key, f.value)
	}
}

// convertFields 批量转换 Field 切片为 zap.Field 切片
func convertFields(fields []Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}

	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = field.toZapField()
	}
	return zapFields
}

// Field 构造函数 - 提供便捷的字段创建方式

// String 创建字符串字段
func String(key, value string) Field {
	return &logField{
		key:       key,
		value:     value,
		fieldType: StringFieldType,
	}
}

// Int 创建整数字段
func Int(key string, value int) Field {
	return &logField{
		key:       key,
		value:     value,
		fieldType: IntFieldType,
	}
}

// Int64 创建 int64 字段
func Int64(key string, value int64) Field {
	return &logField{
		key:       key,
		value:     value,
		fieldType: Int64FieldType,
	}
}

// Float64 创建 float64 字段
func Float64(key string, value float64) Field {
	return &logField{
		key:       key,
		value:     value,
		fieldType: Float64FieldType,
	}
}

// Bool 创建布尔字段
func Bool(key string, value bool) Field {
	return &logField{
		key:       key,
		value:     value,
		fieldType: BoolFieldType,
	}
}

// ErrorField 创建错误字段
func ErrorField(err error) Field {
	return &logField{
		key:       "error",
		value:     err,
		fieldType: ErrorFieldType,
	}
}

// NamedError 创建带自定义键名的错误字段
func NamedError(key string, err error) Field {
	return &logField{
		key:       key,
		value:     err,
		fieldType: ErrorFieldType,
	}
}

// Time 创建时间字段
func Time(key string, value time.Time) Field {
	return &logField{
		key:       key,
		value:     value,
		fieldType: TimeFieldType,
	}
}

// Duration 创建时间段字段
func Duration(key string, value time.Duration) Field {
	return &logField{
		key:       key,
		value:     value,
		fieldType: DurationFieldType,
	}
}

// Binary 创建二进制字段
func Binary(key string, value []byte) Field {
	return &logField{
		key:       key,
		value:     value,
		fieldType: BinaryFieldType,
	}
}

// ByteString 创建字节字符串字段
func ByteString(key string, value []byte) Field {
	return &logField{
		key:       key,
		value:     value,
		fieldType: ByteStringFieldType,
	}
}

// Any 创建任意类型字段
func Any(key string, value interface{}) Field {
	return &logField{
		key:       key,
		value:     value,
		fieldType: AnyFieldType,
	}
}

// Skip 创建跳过字段
func Skip() Field {
	return &logField{
		fieldType: SkipFieldType,
	}
}

// 向后兼容函数 - 如果有现有代码使用 zap.Field，可以通过这些函数迁移

// FromZapField 从 zap.Field 创建 Field（用于迁移）
func FromZapField(zapField zap.Field) Field {
	// 根据 zap.Field 的类型创建对应的 Field
	switch zapField.Type {
	case zapcore.StringType:
		return String(zapField.Key, zapField.String)
	case zapcore.Int64Type:
		return Int64(zapField.Key, zapField.Integer)
	case zapcore.Float64Type:
		return Float64(zapField.Key, math.Float64frombits(uint64(zapField.Integer)))
	case zapcore.BoolType:
		return Bool(zapField.Key, zapField.Integer == 1)
	case zapcore.ErrorType:
		if err, ok := zapField.Interface.(error); ok {
			return ErrorField(err)
		}
		return Any(zapField.Key, zapField.Interface)
	case zapcore.TimeType:
		if zapField.Interface != nil {
			if t, ok := zapField.Interface.(time.Time); ok {
				return Time(zapField.Key, t)
			}
		}
		return Any(zapField.Key, zapField.Interface)
	case zapcore.DurationType:
		return Duration(zapField.Key, time.Duration(zapField.Integer))
	case zapcore.BinaryType:
		if data, ok := zapField.Interface.([]byte); ok {
			return Binary(zapField.Key, data)
		}
		return Any(zapField.Key, zapField.Interface)
	case zapcore.ByteStringType:
		if data, ok := zapField.Interface.([]byte); ok {
			return ByteString(zapField.Key, data)
		}
		return Any(zapField.Key, zapField.Interface)
	case zapcore.SkipType:
		return Skip()
	default:
		return Any(zapField.Key, zapField.Interface)
	}
}

// FromZapFields 批量转换 zap.Field 切片（用于迁移）
func FromZapFields(zapFields []zap.Field) []Field {
	fields := make([]Field, len(zapFields))
	for i, zapField := range zapFields {
		fields[i] = FromZapField(zapField)
	}
	return fields
}

// Logger wraps zap logger
type Logger struct {
	log   *zap.Logger
	async bool
	pool  *concurrent.Pool[struct{}]
}

var (
	defaultLogger *Logger
	defaultOnce   sync.Once
)

// Option represents an option for configuring the logger
type Option func(*Logger)

// WithAsync enables async logging with specified worker and buffer size
func WithAsync(workers, bufferSize int) Option {
	if workers <= 0 {
		workers = defaultAsyncWorkers
	}
	if bufferSize <= 0 {
		bufferSize = defaultAsyncBuffer
	}

	return func(l *Logger) {
		l.async = true
		l.pool = concurrent.NewPool[struct{}](workers, bufferSize, concurrent.WithErrorHandler[struct{}](func(err error) {
			if l.log != nil {
				l.log.Error("Async logger error",
					zap.Error(err),
					zap.Int("workers", workers),
					zap.Int("buffer_size", bufferSize))
			}
		}))
		l.pool.Start()
	}
}

// stdoutWriteSyncer 包装 bufio.Writer 实现 WriteSyncer 接口
type stdoutWriteSyncer struct {
	writer *bufio.Writer
	mu     sync.Mutex
}

// newStdoutWriteSyncer 创建一个新的 stdoutWriteSyncer
func newStdoutWriteSyncer(w *bufio.Writer) *stdoutWriteSyncer {
	return &stdoutWriteSyncer{
		writer: w,
	}
}

func (ws *stdoutWriteSyncer) Write(p []byte) (n int, err error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	n, err = ws.writer.Write(p)
	if err != nil {
		return n, err
	}

	// 确保所有数据都写入
	for n < len(p) {
		var m int
		m, err = ws.writer.Write(p[n:])
		if err != nil {
			return n, err
		}
		n += m
	}

	// 立即刷新缓冲区，确保日志及时显示
	err = ws.writer.Flush()
	return n, err
}

func (ws *stdoutWriteSyncer) Sync() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.writer.Flush()
}

// getEncoderConfig returns a new encoder configuration
func getEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// getLogWriter returns a new log writer with rotation
func getLogWriter(cfg *config.LogConfig) *lumberjack.Logger {
	// 使用默认值填充未设置的配置
	maxSize := cfg.MaxSize
	if maxSize <= 0 {
		maxSize = defaultMaxSize
	}
	maxBackups := cfg.MaxBackups
	if maxBackups <= 0 {
		maxBackups = defaultMaxBackups
	}
	maxAge := cfg.MaxAge
	if maxAge <= 0 {
		maxAge = defaultMaxAge
	}

	return &lumberjack.Logger{
		Filename:   filepath.Join(cfg.Directory, strings.Replace(cfg.Filename, "%Y-%m-%d", time.Now().Format("2006-01-02"), -1)),
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   cfg.Compress,
	}
}

// getConsoleWriter 获取控制台输出的 WriteSyncer
func getConsoleWriter() zapcore.WriteSyncer {
	if runtime.GOOS == "windows" {
		writer := bufio.NewWriterSize(os.Stdout, 4096) // 增加缓冲区大小
		return zapcore.AddSync(newStdoutWriteSyncer(writer))
	}
	return zapcore.AddSync(os.Stdout)
}

// NewLogger creates a new logger instance
func NewLogger(cfg *config.LogConfig, opts ...Option) (*Logger, error) {
	if cfg == nil {
		return nil, fmt.Errorf("logger config cannot be nil")
	}

	// Ensure logs directory exists
	if err := os.MkdirAll(cfg.Directory, 0755); err != nil {
		return nil, fmt.Errorf("can't create log directory: %w", err)
	}

	// Parse log level with default fallback
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil || cfg.Level == "" {
		level = zapcore.InfoLevel
	}

	cores := make([]zapcore.Core, 0, 2) // 预分配容量

	// File output
	cores = append(cores, zapcore.NewCore(
		zapcore.NewJSONEncoder(getEncoderConfig()),
		zapcore.AddSync(getLogWriter(cfg)),
		level,
	))

	// Console output
	if cfg.Console {
		cores = append(cores, zapcore.NewCore(
			zapcore.NewConsoleEncoder(getEncoderConfig()),
			getConsoleWriter(),
			level,
		))
	}

	// Create logger
	zapLogger := zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	logger := &Logger{log: zapLogger}

	// Apply options
	for _, opt := range opts {
		opt(logger)
	}

	return logger, nil
}

// asyncLog 异步记录日志
func (l *Logger) asyncLog(level zapcore.Level, msg string, fields ...Field) {
	if l == nil || l.pool == nil {
		return
	}

	// 在提交异步任务前捕获调用者信息
	caller := zapcore.NewEntryCaller(runtime.Caller(2))

	// 转换字段
	zapFields := convertFields(fields)

	// 创建日志任务
	task := func(ctx context.Context) (struct{}, error) {
		if l.log == nil {
			return struct{}{}, fmt.Errorf("logger is nil")
		}

		// 创建一个新的 Entry
		entry := zapcore.Entry{
			Level:      level,
			Time:       time.Now(),
			Message:    msg,
			Caller:     caller,
			LoggerName: "",
		}

		// 使用 log.Core().Write 直接写入，以保持正确的调用位置
		if err := l.log.Core().Write(entry, zapFields); err != nil {
			return struct{}{}, err
		}

		return struct{}{}, nil
	}

	// 尝试提交任务
	if err := l.pool.Submit(task); err != nil {
		// 如果工作池已满或出错，直接同步写入
		if l.log != nil {
			entry := zapcore.Entry{
				Level:      level,
				Time:       time.Now(),
				Message:    msg,
				Caller:     caller,
				LoggerName: "",
			}
			_ = l.log.Core().Write(entry, zapFields)
		}
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...Field) {
	if l == nil || l.log == nil {
		return
	}
	if l.async {
		l.asyncLog(zapcore.DebugLevel, msg, fields...)
	} else {
		l.log.Debug(msg, convertFields(fields)...)
	}
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...Field) {
	if l == nil || l.log == nil {
		return
	}
	if l.async {
		l.asyncLog(zapcore.InfoLevel, msg, fields...)
	} else {
		l.log.Info(msg, convertFields(fields)...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...Field) {
	if l == nil || l.log == nil {
		return
	}
	if l.async {
		l.asyncLog(zapcore.WarnLevel, msg, fields...)
	} else {
		l.log.Warn(msg, convertFields(fields)...)
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...Field) {
	if l == nil || l.log == nil {
		return
	}
	if l.async {
		l.asyncLog(zapcore.ErrorLevel, msg, fields...)
	} else {
		l.log.Error(msg, convertFields(fields)...)
	}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, fields ...Field) {
	if l == nil || l.log == nil {
		return
	}
	zapFields := convertFields(fields)
	if l.async {
		// 对于Fatal级别，直接同步写入
		l.log.Fatal(msg, zapFields...)
	} else {
		l.log.Fatal(msg, zapFields...)
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	if l == nil || l.log == nil {
		return nil
	}

	if l.async && l.pool != nil {
		l.pool.Stop() // 停止工作池
	}

	return l.log.Sync()
}

// Close properly closes the logger
func (l *Logger) Close() error {
	return l.Sync()
}

// InitLogger initializes the default logger
func InitLogger(cfg *config.LogConfig, opts ...Option) error {
	var err error
	defaultOnce.Do(func() {
		var logger *Logger
		logger, err = NewLogger(cfg, opts...)
		if err == nil {
			defaultLogger = logger
		}
	})
	return err
}

// With creates a child logger with additional fields
func (l *Logger) With(fields ...Field) *Logger {
	if l == nil || l.log == nil {
		return &Logger{}
	}

	zapFields := convertFields(fields)
	childLogger := &Logger{
		log:   l.log.With(zapFields...),
		async: l.async,
	}

	if l.async {
		childLogger.pool = l.pool // 复用父logger的工作池
	}

	return childLogger
}

// GetLogger returns the default logger instance
func GetLogger() *Logger {
	return defaultLogger
}

// Debug  logger methods
func Debug(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Debug(msg, fields...)
	}
}

func Info(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Warn(msg, fields...)
	}
}

func ErrorLog(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Error(msg, fields...)
	}
}

func Fatal(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Fatal(msg, fields...)
	}
}

func With(fields ...Field) *Logger {
	if defaultLogger != nil {
		return defaultLogger.With(fields...)
	}
	return &Logger{}
}

// WithMongoDB 添加 MongoDB 日志支持
func WithMongoDB(mongoDB *database.MongoDB, collection string) Option {
	if collection == "" {
		collection = "logs"
	}
	return func(l *Logger) {
		mongoCore := NewMongoCore(mongoDB, collection)
		l.log = l.log.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(core, mongoCore)
		}))
	}
}
