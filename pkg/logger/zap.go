package logger

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"fiber_web/pkg/config"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap logger
type Logger struct {
	log *zap.Logger
}

var defaultLogger *Logger

// NewLogger creates a new logger instance
func NewLogger(cfg *config.LogConfig) (*Logger, error) {
	// Ensure logs directory exists
	if err := os.MkdirAll(cfg.Directory, 0755); err != nil {
		return nil, fmt.Errorf("can't create log directory: %w", err)
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
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

	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Set log file path
	currentTime := time.Now()
	logFile := filepath.Join(cfg.Directory,
		strings.Replace(cfg.Filename, "%Y-%m-%d", currentTime.Format("2006-01-02"), -1))

	// Set up lumberjack for log rotation
	logWriter := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	// Create cores
	var cores []zapcore.Core

	// File output
	cores = append(cores, zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(logWriter),
		level,
	))

	// Console output
	if cfg.Console {
		var consoleWriter zapcore.WriteSyncer
		if runtime.GOOS == "windows" {
			// Windows 平台使用 bufio.Writer 包装 stdout
			stdout := bufio.NewWriter(os.Stdout)
			consoleWriter = zapcore.AddSync(stdoutWriteSyncer{stdout})
		} else {
			consoleWriter = zapcore.AddSync(os.Stdout)
		}

		cores = append(cores, zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			consoleWriter,
			level,
		))
	}

	// Create logger
	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{log: logger}, nil
}

// stdoutWriteSyncer 包装 bufio.Writer 实现 WriteSyncer 接口
type stdoutWriteSyncer struct {
	*bufio.Writer
}

func (ws stdoutWriteSyncer) Sync() error {
	return ws.Flush()
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	if l == nil || l.log == nil {
		return nil
	}
	return l.log.Sync()
}

// InitLogger initializes the default logger
func InitLogger(cfg *config.LogConfig) error {
	logger, err := NewLogger(cfg)
	if err != nil {
		return err
	}
	defaultLogger = logger
	return nil
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.log.Debug(msg, fields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.log.Info(msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.log.Warn(msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.log.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.log.Fatal(msg, fields...)
}

// With creates a child logger with additional fields
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{log: l.log.With(fields...)}
}

// Debug Default logger methods
func Debug(msg string, fields ...zap.Field) {
	defaultLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	defaultLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	defaultLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	defaultLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	defaultLogger.Fatal(msg, fields...)
}

func With(fields ...zap.Field) *Logger {
	return defaultLogger.With(fields...)
}

func Sync() error {
	return defaultLogger.Sync()
}

// GetLogger 获取默认logger实例
func GetLogger() *zap.Logger {
	return defaultLogger.log
}

// GetDefaultLogger 获取默认Logger实例
func GetDefaultLogger() *Logger {
	return defaultLogger
}
