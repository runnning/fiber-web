package repository

import (
	"fiber_web/pkg/redis"

	"gorm.io/gorm"
)

// InitRepositories 初始化所有仓储
func InitRepositories(db *gorm.DB, redisClient *redis.Client) *Repositories {
	return &Repositories{
		UserRepository: NewUserRepository(db, redisClient),
		// 在这里添加其他仓储的初始化
		AdminUserRepository: NewAdminUserRepository(db, redisClient),
		ApiRepository:       NewApiRepository(db, redisClient),
		MenuRepository:      NewMenuRepository(db, redisClient),
	}
}

// Repositories 仓储集合
type Repositories struct {
	UserRepository UserRepository
	// 在这里添加其他仓储
	AdminUserRepository AdminUserRepository
	ApiRepository       ApiRepository
	MenuRepository      MenuRepository
	RoleRepository      RoleRepository
}
