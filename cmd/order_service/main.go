package main

import (
	"context"

	"github.com/stsolovey/order_tracker/internal/config"
	"github.com/stsolovey/order_tracker/internal/logger"
	"github.com/stsolovey/order_tracker/internal/storage"
)

func main() {
	log := logger.New()

	cfg, err := config.New(log, "./.env")
	if err != nil {
		log.WithError(err).Panic("Failed to initialize config")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storageSystem, err := storage.NewStorage(ctx, cfg.DatabaseURL)
	if err != nil {
		log.WithError(err).Panic("Failed to initialize storage")
	}

	if err := storageSystem.Migrate(log); err != nil {
		log.WithError(err).Panic("Failed to execute migrations")
	}

}
