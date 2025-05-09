package bootstrap

import (
	"context"
	"log"
	"sync"
	"time"
)

// 生命周期阶段常量
const (
	HookBeforeInit = "beforeInit"
	HookAfterInit  = "afterInit"
	HookBeforeStop = "beforeStop"
	HookAfterStop  = "afterStop"
)

// LifecycleHook 生命周期钩子函数类型
type LifecycleHook func(ctx context.Context) error

// Component 组件接口
type Component interface {
	Init(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Bootstrapper 管理应用程序初始化和生命周期
type Bootstrapper struct {
	components []Component
	hooks      struct {
		beforeInit []LifecycleHook
		afterInit  []LifecycleHook
		beforeStop []LifecycleHook
		afterStop  []LifecycleHook
	}
	mu sync.Mutex
}

// New 创建新的 Bootstrapper
func New() *Bootstrapper {
	return &Bootstrapper{
		components: make([]Component, 0),
		hooks: struct {
			beforeInit []LifecycleHook
			afterInit  []LifecycleHook
			beforeStop []LifecycleHook
			afterStop  []LifecycleHook
		}{
			beforeInit: make([]LifecycleHook, 0),
			afterInit:  make([]LifecycleHook, 0),
			beforeStop: make([]LifecycleHook, 0),
			afterStop:  make([]LifecycleHook, 0),
		},
	}
}

// AddComponent 添加组件
func (b *Bootstrapper) AddComponent(c Component) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.components = append(b.components, c)
}

// AddHook 添加生命周期钩子
func (b *Bootstrapper) AddHook(phase string, hook LifecycleHook) {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch phase {
	case HookBeforeInit:
		b.hooks.beforeInit = append(b.hooks.beforeInit, hook)
	case HookAfterInit:
		b.hooks.afterInit = append(b.hooks.afterInit, hook)
	case HookBeforeStop:
		b.hooks.beforeStop = append(b.hooks.beforeStop, hook)
	case HookAfterStop:
		b.hooks.afterStop = append(b.hooks.afterStop, hook)
	}
}

// Bootstrap 运行初始化流程
func (b *Bootstrapper) Bootstrap(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 执行初始化前钩子
	for _, hook := range b.hooks.beforeInit {
		if err := hook(ctx); err != nil {
			return err
		}
	}

	// 初始化组件
	for _, component := range b.components {
		if err := component.Init(ctx); err != nil {
			return err
		}
	}

	// 执行初始化后钩子
	for _, hook := range b.hooks.afterInit {
		if err := hook(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Start 启动所有组件
func (b *Bootstrapper) Start(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, component := range b.components {
		if err := component.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Shutdown 运行关闭流程
func (b *Bootstrapper) Shutdown() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 快速执行停止前钩子
	for _, hook := range b.hooks.beforeStop {
		hookCtx, hookCancel := context.WithTimeout(ctx, 1*time.Second)
		if err := hook(hookCtx); err != nil {
			log.Printf("钩子执行出错: %v\n", err)
		}
		hookCancel()
	}

	// 并发停止所有组件
	var wg sync.WaitGroup
	for i := len(b.components) - 1; i >= 0; i-- {
		wg.Add(1)
		go func(component Component) {
			defer wg.Done()
			if err := component.Stop(ctx); err != nil {
				log.Printf("组件停止出错: %v\n", err)
			}
		}(b.components[i])
	}

	// 等待所有组件停止，但最多等待3秒
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 继续执行
	case <-time.After(3 * time.Second):
		log.Println("组件关闭超时")
	}

	// 快速执行停止后钩子
	for _, hook := range b.hooks.afterStop {
		hookCtx, hookCancel := context.WithTimeout(ctx, 1*time.Second)
		if err := hook(hookCtx); err != nil {
			log.Printf("钩子执行出错: %v\n", err)
		}
		hookCancel()
	}

	return nil
}
