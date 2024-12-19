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
	boot.AddHook(bootstrap.HookBeforeInit, func(ctx context.Context) error {
		log.Printf("[%s] Starting application initialization...", appName)
		return nil
	})

	boot.AddHook(bootstrap.HookAfterInit, func(ctx context.Context) error {
		log.Printf("[%s] Application initialization completed", appName)
		return nil
	})

	boot.AddHook(bootstrap.HookBeforeStop, func(ctx context.Context) error {
		log.Printf("[%s] Starting application shutdown...", appName)
		return nil
	})

	boot.AddHook(bootstrap.HookAfterStop, func(ctx context.Context) error {
		log.Printf("[%s] Application shutdown completed", appName)
		return nil
	})
}

// APILifecycleHooks API服务特定的生命周期钩子
type APILifecycleHooks struct {
	BaseLifecycleHooks
}

func (h *APILifecycleHooks) RegisterHooks(boot *bootstrap.Bootstrapper, appName AppType) {
	h.BaseLifecycleHooks.RegisterHooks(boot, appName)
	boot.AddHook(bootstrap.HookAfterInit, func(ctx context.Context) error {
		log.Printf("[%s] API specific initialization completed", appName)
		return nil
	})
}

// AdminLifecycleHooks Admin服务特定的生命周期钩子
type AdminLifecycleHooks struct {
	BaseLifecycleHooks
}

func (h *AdminLifecycleHooks) RegisterHooks(boot *bootstrap.Bootstrapper, appName AppType) {
	h.BaseLifecycleHooks.RegisterHooks(boot, appName)
	boot.AddHook(bootstrap.HookAfterInit, func(ctx context.Context) error {
		log.Printf("[%s] Admin specific initialization completed", appName)
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
