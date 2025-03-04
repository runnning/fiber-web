package logger

import (
	"context"
	"time"

	"fiber_web/pkg/database"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap/zapcore"
)

type LogEntry struct {
	Timestamp time.Time              `bson:"timestamp"`
	Level     string                 `bson:"level"`
	Message   string                 `bson:"message"`
	Caller    string                 `bson:"caller"`
	Fields    map[string]interface{} `bson:"fields,omitempty"`
}

type mongoCore struct {
	collection *mongo.Collection
}

func NewMongoCore(mongoDB *database.MongoDB, collectionName string) zapcore.Core {
	return &mongoCore{
		collection: mongoDB.Collection(collectionName),
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

func (c *mongoCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	logEntry := LogEntry{
		Timestamp: entry.Time,
		Level:     entry.Level.String(),
		Message:   entry.Message,
		Caller:    entry.Caller.String(),
	}

	if len(fields) > 0 {
		encoder := zapcore.NewMapObjectEncoder()
		for _, f := range fields {
			f.AddTo(encoder)
		}
		logEntry.Fields = encoder.Fields
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.collection.InsertOne(ctx, logEntry)
	if err != nil {
		return err
	}
	return err
}

func (c *mongoCore) Sync() error {
	return nil
}
