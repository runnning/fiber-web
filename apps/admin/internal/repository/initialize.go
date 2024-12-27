package repository

import (
	"fiber_web/pkg/redis"
	"gorm.io/gorm"
)

// InitRepositories 初始化所有仓储
func InitRepositories(db *gorm.DB, redisClient *redis.Client) *Repositories {
	return &Repositories{
		User: NewUserRepository(db, redisClient),
		// 在这里添加其他仓储的初始化
	}
}

// Repositories 仓储集合
type Repositories struct {
	User UserRepository
	// 在这里添加其他仓储
}
