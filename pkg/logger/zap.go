package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

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
func NewLogger(env string) (*Logger, error) {
	var config zap.Config

	// Ensure logs directory exists
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, fmt.Errorf("can't create log directory: %v", err)
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

	// Set different configs based on environment
	if env == "development" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	config.EncoderConfig = encoderConfig

	// Create core with file and console output
	currentTime := time.Now().Format("2006-01-02")
	logFile := filepath.Join("logs", fmt.Sprintf("%s.log", currentTime))

	// Set up lumberjack for log rotation
	logWriter := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    1,    // megabytes
		MaxBackups: 3,    // number of backups
		MaxAge:     28,   // days
		Compress:   true, // compress the backups
	}

	// Open log files
	// file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	return nil, fmt.Errorf("can't open log file: %v", err)
	// }

	// Create cores
	cores := []zapcore.Core{
		zapcore.NewCore(
			zapcore.NewJSONEncoder(config.EncoderConfig),
			zapcore.AddSync(logWriter),
			config.Level,
		),
		// zapcore.NewCore(
		// 	zapcore.NewJSONEncoder(config.EncoderConfig),
		// 	zapcore.AddSync(file),
		// 	zapcore.DebugLevel,
		// ),
	}

	// Add console output in development
	if env == "development" {
		cores = append(cores, zapcore.NewCore(
			zapcore.NewConsoleEncoder(config.EncoderConfig),
			zapcore.AddSync(os.Stdout),
			config.Level,
		))
	}

	// Create logger
	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	return &Logger{log: logger}, nil
}

// InitLogger initializes the default logger
func InitLogger(env string) error {
	logger, err := NewLogger(env)
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

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.log.Sync()
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
