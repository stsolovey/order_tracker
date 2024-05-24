package main

import (
	"context"

	_ "github.com/jackc/pgx/v5/stdlib" // Importing `pgx/v5/stdlib` is necessary for `sql.Open("pgx", s.dsn)`.
	"github.com/stsolovey/order_tracker/internal/config"
	"github.com/stsolovey/order_tracker/internal/logger"
	"github.com/stsolovey/order_tracker/internal/storage"
)

func main() {
	cfg := config.New("./.env")
	log := logger.New(cfg.LogLevel)

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
