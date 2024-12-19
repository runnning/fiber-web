package bootstrap

import (
	"context"
	"sync"
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
	initFuncs  []func(context.Context) error
	closeFuncs []func() error
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
		initFuncs:  make([]func(context.Context) error, 0),
		closeFuncs: make([]func() error, 0),
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

// Register 注册初始化和关闭函数
func (b *Bootstrapper) Register(initFn func(context.Context) error, closeFn func() error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if initFn != nil {
		b.initFuncs = append(b.initFuncs, initFn)
	}
	if closeFn != nil {
		b.closeFuncs = append(b.closeFuncs, closeFn)
	}
}

// AddHook 添加生命周期钩子
func (b *Bootstrapper) AddHook(phase string, hook LifecycleHook) {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch phase {
	case "beforeInit":
		b.hooks.beforeInit = append(b.hooks.beforeInit, hook)
	case "afterInit":
		b.hooks.afterInit = append(b.hooks.afterInit, hook)
	case "beforeStop":
		b.hooks.beforeStop = append(b.hooks.beforeStop, hook)
	case "afterStop":
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

	// 执行注册的初始化函数
	for _, fn := range b.initFuncs {
		if err := fn(ctx); err != nil {
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

	ctx := context.Background()

	// 执行停止前钩子
	for _, hook := range b.hooks.beforeStop {
		if err := hook(ctx); err != nil {
			return err
		}
	}

	// 逆序停止组件
	for i := len(b.components) - 1; i >= 0; i-- {
		if err := b.components[i].Stop(ctx); err != nil {
			return err
		}
	}

	// 逆序执行关闭函数
	for i := len(b.closeFuncs) - 1; i >= 0; i-- {
		if err := b.closeFuncs[i](); err != nil {
			return err
		}
	}

	// 执行停止后钩子
	for _, hook := range b.hooks.afterStop {
		if err := hook(ctx); err != nil {
			return err
		}
	}

	return nil
}

// 生命周期阶段常量
const (
	HookBeforeInit = "beforeInit"
	HookAfterInit  = "afterInit"
	HookBeforeStop = "beforeStop"
	HookAfterStop  = "afterStop"
)
