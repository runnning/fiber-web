package database

import (
	"fiber_web/pkg/config"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// DBManager 管理多个数据库连接
type DBManager struct {
	dbs map[string]*Database
}

// Database 包装单个数据库连接
type Database struct {
	db *gorm.DB
}

// NewDBManager 创建数据库管理器
func NewDBManager(cfg *config.DatabaseConfig) (*DBManager, error) {
	manager := &DBManager{
		dbs: make(map[string]*Database),
	}

	if cfg.MultiDB {
		// 多库模式
		for name, dbConfig := range cfg.Databases {
			db, err := newDatabase(&dbConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to create database %s: %w", name, err)
			}
			manager.dbs[name] = db
		}
	} else {
		// 单库模式
		db, err := newDatabase(&cfg.Default)
		if err != nil {
			return nil, err
		}
		manager.dbs["default"] = db
	}

	return manager, nil
}

// newDatabase 创建单个数据库连接
func newDatabase(cfg *config.DBConfig) (*Database, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{db: db}, nil
}

// GetDB 获取指定名称的数据库连接
func (m *DBManager) GetDB(name string) (*Database, error) {
	db, exists := m.dbs[name]
	if !exists {
		return nil, fmt.Errorf("database %s not found", name)
	}
	return db, nil
}

// Close 关闭闭所有数据库连接
func (m *DBManager) Close() error {
	var errs []error
	for name, db := range m.dbs {
		if err := db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close database %s: %w", name, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing databases: %v", errs)
	}
	return nil
}

// DB 方法保持不变
func (d *Database) DB() *gorm.DB {
	return d.db
}

func (d *Database) AutoMigrate(models ...interface{}) error {
	return d.db.AutoMigrate(models...)
}

func (d *Database) Begin() *gorm.DB {
	return d.db.Begin()
}

func (d *Database) Commit() *gorm.DB {
	return d.db.Commit()
}

func (d *Database) Rollback() *gorm.DB {
	return d.db.Rollback()
}

func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
