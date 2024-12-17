package database

import (
	"fiber_web/pkg/config"
	"fiber_web/pkg/logger"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Database wraps the database connection
type Database struct {
	db *gorm.DB
}

// NewMySQL creates a new MySQL database connection
func NewMySQL(cfg *config.Config) (*Database, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)

	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("Failed to get database instance", zap.Error(err))
		return nil, err
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		logger.Error("Failed to ping database", zap.Error(err))
		return nil, err
	}

	logger.Info("Successfully connected to database")
	return &Database{db: db}, nil
}

// DB returns the underlying database connection
func (d *Database) DB() *gorm.DB {
	return d.db
}

// AutoMigrate runs auto migration for given models
func (d *Database) AutoMigrate(models ...interface{}) error {
	return d.db.AutoMigrate(models...)
}

// Begin starts a new transaction
func (d *Database) Begin() *gorm.DB {
	return d.db.Begin()
}

// Commit commits the current transaction
func (d *Database) Commit() *gorm.DB {
	return d.db.Commit()
}

// Rollback rollbacks the current transaction
func (d *Database) Rollback() *gorm.DB {
	return d.db.Rollback()
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
