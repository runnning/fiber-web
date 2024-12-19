package initialize

import (
	"context"
	"fiber_web/apps/admin/internal/bootstrap"
	"log"
)

// LifecycleHooks 定义生命周期钩子接口
type LifecycleHooks interface {
	RegisterHooks(boot *bootstrap.Bootstrapper, appName AppType)
}

// BaseLifecycleHooks 基础生命周期钩子实现
type BaseLifecycleHooks struct{}

func (h *BaseLifecycleHooks) RegisterHooks(boot *bootstrap.Bootstrapper, appName AppType) {
	// 1. 初始化前
	boot.AddHook(bootstrap.HookBeforeInit, func(ctx context.Context) error {
		log.Printf("[%s] Base initialization starting...", appName)
		return nil
	})

	// 2. 初始化后
	boot.AddHook(bootstrap.HookAfterInit, func(ctx context.Context) error {
		log.Printf("[%s] Base initialization completed", appName)
		return nil
	})

	// 3. 停止前
	boot.AddHook(bootstrap.HookBeforeStop, func(ctx context.Context) error {
		log.Printf("[%s] Base shutdown starting...", appName)
		return nil
	})

	// 4. 停止后
	boot.AddHook(bootstrap.HookAfterStop, func(ctx context.Context) error {
		log.Printf("[%s] Base shutdown completed", appName)
		return nil
	})
}

// APILifecycleHooks API服务特定的生命周期钩子
type APILifecycleHooks struct {
	BaseLifecycleHooks
}

func (h *APILifecycleHooks) RegisterHooks(boot *bootstrap.Bootstrapper, appName AppType) {
	// 1. 先注册基础钩子
	h.BaseLifecycleHooks.RegisterHooks(boot, appName)

	// 2. API特定的初始化前钩子
	boot.AddHook(bootstrap.HookBeforeInit, func(ctx context.Context) error {
		log.Printf("[%s] API initialization starting...", appName)
		return nil
	})

	// 3. API特定的初始化后钩子
	boot.AddHook(bootstrap.HookAfterInit, func(ctx context.Context) error {
		log.Printf("[%s] API initialization completed", appName)
		return nil
	})

	// 4. API特定的停止前钩子
	boot.AddHook(bootstrap.HookBeforeStop, func(ctx context.Context) error {
		log.Printf("[%s] API shutdown starting...", appName)
		return nil
	})

	// 5. API特定的停止后钩子
	boot.AddHook(bootstrap.HookAfterStop, func(ctx context.Context) error {
		log.Printf("[%s] API shutdown completed", appName)
		return nil
	})
}

// AdminLifecycleHooks Admin服务特定的生命周期钩子
type AdminLifecycleHooks struct {
	BaseLifecycleHooks
}

func (h *AdminLifecycleHooks) RegisterHooks(boot *bootstrap.Bootstrapper, appName AppType) {
	// 1. 先注册基础钩子
	h.BaseLifecycleHooks.RegisterHooks(boot, appName)

	// 2. Admin特定的初始化前钩子
	boot.AddHook(bootstrap.HookBeforeInit, func(ctx context.Context) error {
		log.Printf("[%s] Admin initialization starting...", appName)
		return nil
	})

	// 3. Admin特定的初始化后钩子
	boot.AddHook(bootstrap.HookAfterInit, func(ctx context.Context) error {
		log.Printf("[%s] Admin initialization completed", appName)
		return nil
	})

	// 4. Admin特定的停止前钩子
	boot.AddHook(bootstrap.HookBeforeStop, func(ctx context.Context) error {
		log.Printf("[%s] Admin shutdown starting...", appName)
		return nil
	})

	// 5. Admin特定的停止后钩子
	boot.AddHook(bootstrap.HookAfterStop, func(ctx context.Context) error {
		log.Printf("[%s] Admin shutdown completed", appName)
		return nil
	})
}

// NewLifecycleHooks 创建对应类型的生命周期钩子
func NewLifecycleHooks(appType AppType) LifecycleHooks {
	switch appType {
	case AppTypeAPI:
		return &APILifecycleHooks{}
	case AppTypeAdmin:
		return &AdminLifecycleHooks{}
	default:
		return &BaseLifecycleHooks{}
	}
}
