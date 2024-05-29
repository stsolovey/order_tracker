package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib" // Importing `pgx/v5/stdlib` is necessary for `sql.Open("pgx", s.dsn)`.
	"github.com/nats-io/nats.go"
	"github.com/stsolovey/order_tracker/internal/config"
	"github.com/stsolovey/order_tracker/internal/logger"
	natsconsumer "github.com/stsolovey/order_tracker/internal/nats-consumer"
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

	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		log.WithError(err).Panic("Failed to connect to NATS")
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		log.WithError(err).Panic("Failed to get JetStream context")
	}

	streamName := "ORDERS"
	streamSubjects := []string{"orders"}
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: streamSubjects,
	})

	if err != nil && !errors.Is(err, nats.ErrStreamNameAlreadyInUse) {
		log.WithError(err).Panic("Failed to create stream")
	}

	natsClient, err := natsconsumer.New(cfg, log, app)
	if err != nil {
		log.WithError(err).Panic("Failed to initialize NATS client")
	}
	defer natsClient.Close()

	if err := natsClient.Subscribe(ctx, "orders"); err != nil {
		log.WithError(err).Panic("Failed to subscribe to NATS subject")
	}

	<-ctx.Done() // bad solution for "listening" (without graceful shutdown)
}
