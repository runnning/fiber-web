package logger

import (
	"context"
	"time"

	"fiber_web/pkg/database"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap/zapcore"
)

// LogEntry MongoDB日志条目结构
type LogEntry struct {
	Timestamp time.Time      `bson:"timestamp"`
	Level     string         `bson:"level"`
	Message   string         `bson:"message"`
	Caller    string         `bson:"caller"`
	Fields    map[string]any `bson:"fields,omitempty"`
}

// mongoCore MongoDB日志核心实现
type mongoCore struct {
	collection *mongo.Collection
	encoder    zapcore.Encoder
}

// NewMongoCore 创建MongoDB日志核心
func NewMongoCore(mongoDB *database.MongoDB, collectionName string) zapcore.Core {
	return &mongoCore{
		collection: mongoDB.Collection(collectionName),
		encoder:    zapcore.NewJSONEncoder(zapcore.EncoderConfig{}),
	}
}

func (c *mongoCore) Enabled(level zapcore.Level) bool {
	return true
}

func (c *mongoCore) With(fields []zapcore.Field) zapcore.Core {
	return c
}

func (c *mongoCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(entry, c)
}

// processFields 处理日志字段
func (c *mongoCore) processFields(fields []zapcore.Field) map[string]any {
	if len(fields) == 0 {
		return nil
	}

	// 使用 MapObjectEncoder 处理字段
	encoder := zapcore.NewMapObjectEncoder()
	for _, f := range fields {
		f.AddTo(encoder)
	}

	// 创建新的 map 并复制字段
	result := make(map[string]any, len(encoder.Fields))
	for k, v := range encoder.Fields {
		// 对特殊类型进行处理
		switch val := v.(type) {
		case []byte:
			// 复制字节切片
			newVal := make([]byte, len(val))
			copy(newVal, val)
			result[k] = newVal
		case error:
			// 错误转换为字符串
			result[k] = val.Error()
		default:
			// 其他类型直接赋值
			result[k] = v
		}
	}

	return result
}

func (c *mongoCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	logEntry := LogEntry{
		Timestamp: entry.Time,
		Level:     entry.Level.String(),
		Message:   entry.Message,
		Caller:    entry.Caller.String(),
		Fields:    c.processFields(fields),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.collection.InsertOne(ctx, logEntry)
	return err
}

func (c *mongoCore) Sync() error {
	return nil
}
