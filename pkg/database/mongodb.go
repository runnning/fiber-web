package database

import (
	"context"
	"fiber_web/pkg/config"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoManager 管理多个MongoDB连接
type MongoManager struct {
	dbs map[string]*MongoDB
}

// MongoDB 包装单个MongoDB连接
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewMongoManager 创建MongoDB管理器
func NewMongoManager(cfg *config.MongoDBConfig) (*MongoManager, error) {
	manager := &MongoManager{
		dbs: make(map[string]*MongoDB),
	}

	if cfg.MultiDB {
		// 多库模式
		for name, dbConfig := range cfg.Databases {
			db, err := newMongoDB(&dbConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to create mongodb %s: %w", name, err)
			}
			manager.dbs[name] = db
		}
	} else {
		// 单库模式
		db, err := newMongoDB(&cfg.Default)
		if err != nil {
			return nil, err
		}
		manager.dbs["default"] = db
	}

	return manager, nil
}

// newMongoDB 创建单个MongoDB连接
func newMongoDB(cfg *config.MongoConfig) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 创建MongoDB客户端选项
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetMaxConnIdleTime(cfg.MaxConnIdleTime)

	// 连接到MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	// 获取数据库实例
	database := client.Database(cfg.Database)

	return &MongoDB{
		client:   client,
		database: database,
	}, nil
}

// GetMongoDB 获取指定名称的MongoDB连接
func (m *MongoManager) GetMongoDB(name string) (*MongoDB, error) {
	db, exists := m.dbs[name]
	if !exists {
		return nil, fmt.Errorf("mongodb %s not found", name)
	}
	return db, nil
}

// Close 关闭所有MongoDB连接
func (m *MongoManager) Close() error {
	var errs []error
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for name, db := range m.dbs {
		if err := db.Close(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to close mongodb %s: %w", name, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing mongodb connections: %v", errs)
	}
	return nil
}

// Client 获取MongoDB客户端
func (m *MongoDB) Client() *mongo.Client {
	return m.client
}

// Database 获取MongoDB数据库实例
func (m *MongoDB) Database() *mongo.Database {
	return m.database
}

// Collection 获取指定集合
func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.database.Collection(name)
}

// Close 关闭MongoDB连接
func (m *MongoDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// WithTransaction 执行事务
func (m *MongoDB) WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	session, err := m.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})
	return err
}
