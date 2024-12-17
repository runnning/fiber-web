package bootstrap

import (
	"context"
	"sync"
)

// Bootstrapper manages application initialization
type Bootstrapper struct {
	initFuncs  []func(context.Context) error
	closeFuncs []func() error
	mu         sync.Mutex
}

// New creates a new Bootstrapper
func New() *Bootstrapper {
	return &Bootstrapper{
		initFuncs:  make([]func(context.Context) error, 0),
		closeFuncs: make([]func() error, 0),
	}
}

// Register adds an initialization function
func (b *Bootstrapper) Register(initFn func(context.Context) error, closeFn func() error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initFuncs = append(b.initFuncs, initFn)
	if closeFn != nil {
		b.closeFuncs = append(b.closeFuncs, closeFn)
	}
}

// Bootstrap runs all initialization functions
func (b *Bootstrapper) Bootstrap(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, fn := range b.initFuncs {
		if err := fn(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Shutdown runs all close functions in reverse order
func (b *Bootstrapper) Shutdown() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i := len(b.closeFuncs) - 1; i >= 0; i-- {
		if err := b.closeFuncs[i](); err != nil {
			return err
		}
	}
	return nil
}
