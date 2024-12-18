package main

import (
	"context"
	"fiber_web/apps/admin/internal/initialize/component"
	"log"

	"fiber_web/pkg/config"
)

func main() {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// 创建并运行应用
	manager := component.NewManager(cfg, component.AppTypeAPI)
	if err := manager.Initialize(ctx); err != nil {
		log.Fatal(err)
	}

	if err := manager.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
