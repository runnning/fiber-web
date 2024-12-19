package main

import (
	"context"
	"fiber_web/apps/admin/internal/initialize"
	"log"

	"fiber_web/pkg/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	component := initialize.NewComponent(cfg, initialize.AppTypeAPI)

	if err := component.Initialize(ctx); err != nil {
		log.Fatal(err)
	}

	if err := component.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
