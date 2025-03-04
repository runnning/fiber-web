package logger

import (
	"bufio"
	"context"
	"fmt"
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
)

// Logger wraps zap logger
type Logger struct {
	log   *zap.Logger
	async bool
	pool  *concurrent.Pool[struct{}]
	once  sync.Once      // 用于确保只执行一次关闭操作
	wg    sync.WaitGroup // 用于等待所有异步日志完成
}

var (
	defaultLogger *Logger
	defaultOnce   sync.Once
)

// Option represents an option for configuring the logger
type Option func(*Logger)

// WithAsync enables async logging with specified worker and buffer size
func WithAsync(workers, bufferSize int) Option {
	return func(l *Logger) {
		l.async = true
		l.pool = concurrent.NewPool[struct{}](workers, bufferSize, concurrent.WithErrorHandler[struct{}](func(err error) {
			// 如果异步日志发生错误，打印到控制台
			fmt.Printf("Async logger error: %v\n", err)
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
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
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
func (l *Logger) asyncLog(level zapcore.Level, msg string, fields ...zap.Field) {
	if l == nil || l.pool == nil {
		return
	}

	l.wg.Add(1)
	if err := l.pool.Submit(func(ctx context.Context) (struct{}, error) {
		defer l.wg.Done()
		if l.log == nil {
			return struct{}{}, fmt.Errorf("logger is nil")
		}

		switch level {
		case zapcore.DebugLevel:
			l.log.Debug(msg, fields...)
		case zapcore.InfoLevel:
			l.log.Info(msg, fields...)
		case zapcore.WarnLevel:
			l.log.Warn(msg, fields...)
		case zapcore.ErrorLevel:
			l.log.Error(msg, fields...)
		case zapcore.FatalLevel:
			l.log.Fatal(msg, fields...)
		default:
			return struct{}{}, fmt.Errorf("unknown log level: %v", level)
		}
		return struct{}{}, nil
	}); err != nil {
		l.wg.Done() // 如果提交失败，需要减少计数
		if l.log != nil {
			l.log.Error("Failed to submit async log", zap.Error(err))
			l.log.Log(level, msg, fields...)
		}
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	if l == nil || l.log == nil {
		return
	}
	if l.async {
		l.asyncLog(zapcore.DebugLevel, msg, fields...)
	} else {
		l.log.Debug(msg, fields...)
	}
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...zap.Field) {
	if l == nil || l.log == nil {
		return
	}
	if l.async {
		l.asyncLog(zapcore.InfoLevel, msg, fields...)
	} else {
		l.log.Info(msg, fields...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	if l == nil || l.log == nil {
		return
	}
	if l.async {
		l.asyncLog(zapcore.WarnLevel, msg, fields...)
	} else {
		l.log.Warn(msg, fields...)
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...zap.Field) {
	if l == nil || l.log == nil {
		return
	}
	if l.async {
		l.asyncLog(zapcore.ErrorLevel, msg, fields...)
	} else {
		l.log.Error(msg, fields...)
	}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	if l == nil || l.log == nil {
		return
	}
	if l.async {
		// 对于Fatal级别，直接同步写入
		l.log.Fatal(msg, fields...)
	} else {
		l.log.Fatal(msg, fields...)
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	if l == nil || l.log == nil {
		return nil
	}

	var err error
	l.once.Do(func() {
		if l.async {
			// 等待所有异步日志完成
			l.wg.Wait()
			// 停止工作池
			l.pool.Stop()
		}
		err = l.log.Sync()
	})
	return err
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
func (l *Logger) With(fields ...zap.Field) *Logger {
	if l == nil || l.log == nil {
		return &Logger{}
	}

	childLogger := &Logger{
		log:   l.log.With(fields...),
		async: l.async,
		//wg:    l.wg,
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
func Debug(msg string, fields ...zap.Field) {
	if defaultLogger != nil {
		defaultLogger.Debug(msg, fields...)
	}
}

func Info(msg string, fields ...zap.Field) {
	if defaultLogger != nil {
		defaultLogger.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	if defaultLogger != nil {
		defaultLogger.Warn(msg, fields...)
	}
}

func Error(msg string, fields ...zap.Field) {
	if defaultLogger != nil {
		defaultLogger.Error(msg, fields...)
	}
}

func Fatal(msg string, fields ...zap.Field) {
	if defaultLogger != nil {
		defaultLogger.Fatal(msg, fields...)
	}
}

func With(fields ...zap.Field) *Logger {
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
