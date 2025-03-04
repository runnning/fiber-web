package query

import (
	"context"
	"fiber_web/pkg/config"
	"fiber_web/pkg/database"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// TestDBConfig 测试数据库配置
var TestDBConfig = &config.DatabaseConfig{
	MultiDB: false,
	Default: config.DBConfig{
		Host:            "localhost",
		Port:            3306,
		User:            "root",
		Password:        "root",
		DBName:          "fiber_web",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
	},
}

// TestMongoConfig 测试MongoDB配置
var TestMongoConfig = &config.MongoDBConfig{
	MultiDB: false,
	Default: config.MongoConfig{
		URI:             "mongodb://localhost:27017",
		Database:        "test_db",
		Username:        "root",
		Password:        "root",
		AuthSource:      "admin",
		MaxPoolSize:     100,
		MinPoolSize:     10,
		MaxConnIdleTime: time.Hour,
	},
}

// TestHelper 测试辅助结构
type TestHelper struct {
	t          *testing.T
	dbManager  *database.DBManager
	mongoMgr   *database.MongoManager
	cleanupFns []func()
}

// NewTestHelper 创建测试辅助实例
func NewTestHelper(t *testing.T) *TestHelper {
	helper := &TestHelper{
		t:          t,
		cleanupFns: make([]func(), 0),
	}

	// 初始化MySQL连接
	dbManager, err := database.NewDBManager(TestDBConfig)
	if err != nil {
		t.Fatalf("Failed to create database manager: %v", err)
	}
	helper.dbManager = dbManager
	helper.cleanupFns = append(helper.cleanupFns, func() {
		if err := dbManager.Close(); err != nil {
			t.Errorf("Failed to close database manager: %v", err)
		}
	})

	// 初始化MongoDB连接
	mongoMgr, err := database.NewMongoManager(TestMongoConfig)
	if err != nil {
		t.Fatalf("Failed to create mongodb manager: %v", err)
	}
	helper.mongoMgr = mongoMgr
	helper.cleanupFns = append(helper.cleanupFns, func() {
		if err := mongoMgr.Close(); err != nil {
			t.Errorf("Failed to close mongodb manager: %v", err)
		}
	})

	return helper
}

// GetDB 获取MySQL数据库连接
func (h *TestHelper) GetDB() *gorm.DB {
	db, err := h.dbManager.GetDB("default")
	if err != nil {
		h.t.Fatalf("Failed to get database: %v", err)
	}
	return db.DB()
}

// GetMongoDB 获取MongoDB数据库连接
func (h *TestHelper) GetMongoDB() *mongo.Database {
	db, err := h.mongoMgr.GetMongoDB("default")
	if err != nil {
		h.t.Fatalf("Failed to get mongodb: %v", err)
	}
	return db.Database()
}

// Cleanup 清理资源
func (h *TestHelper) Cleanup() {
	for _, fn := range h.cleanupFns {
		fn()
	}
}

// WithTestDB 使用测试数据库运行测试
func WithTestDB(t *testing.T, fn func(db *gorm.DB)) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()
	fn(helper.GetDB())
}

// WithTestMongoDB 使用测试MongoDB运行测试
func WithTestMongoDB(t *testing.T, fn func(db *mongo.Database)) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()
	fn(helper.GetMongoDB())
}

// WithTransaction 在事务中运行测试（MySQL）
func (h *TestHelper) WithTransaction(fn func(tx *gorm.DB) error) {
	db := h.GetDB()
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			h.t.Fatalf("Panic in transaction: %v", r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		h.t.Fatalf("Transaction failed: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		h.t.Fatalf("Failed to commit transaction: %v", err)
	}
}

// WithMongoTransaction 在事务中运行测试（MongoDB）
func (h *TestHelper) WithMongoTransaction(fn func(sessCtx mongo.SessionContext) error) {
	db := h.GetMongoDB()
	if err := db.Client().UseSession(context.Background(), func(sessCtx mongo.SessionContext) error {
		if err := sessCtx.StartTransaction(); err != nil {
			return err
		}

		if err := fn(sessCtx); err != nil {
			if abortErr := sessCtx.AbortTransaction(sessCtx); abortErr != nil {
				h.t.Errorf("Failed to abort transaction: %v", abortErr)
			}
			return err
		}

		return sessCtx.CommitTransaction(sessCtx)
	}); err != nil {
		h.t.Fatalf("Transaction failed: %v", err)
	}
}
