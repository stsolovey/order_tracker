package main

import (
	"context"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib" // Importing `pgx/v5/stdlib` is necessary for `sql.Open("pgx", s.dsn)`.
	"github.com/stsolovey/order_tracker/internal/config"
	"github.com/stsolovey/order_tracker/internal/logger"
	ordercache "github.com/stsolovey/order_tracker/internal/order-cache"
	"github.com/stsolovey/order_tracker/internal/service"
	"github.com/stsolovey/order_tracker/internal/storage"
)

func main() {
	cfg := config.New("./.env")
	log := logger.New(cfg.LogLevel)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	db, err := storage.NewStorage(ctx, log, cfg.DatabaseURL)
	if err != nil {
		log.WithError(err).Panic("Failed to initialize storage")
	}

	if err := db.Migrate(); err != nil {
		log.WithError(err).Panic("Failed to execute migrations")
	}

	orderCache := ordercache.New(log)
	app := service.New(log, orderCache, db)

	if err := app.Init(ctx); err != nil {
		log.WithError(err).Panic("Error app initialisation")
	}
}
