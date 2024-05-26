package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

var (
	ErrDBConnectionFailed  = errors.New("database connection failed")
	ErrTimeoutWaitingForDB = errors.New("timeout waiting for DB to be ready")
)

type Storage struct {
	db  *pgxpool.Pool
	dsn string
}

type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

func (s *Storage) DB() *pgxpool.Pool {
	return s.db
}

func NewStorage(ctx context.Context, dsn string) (*Storage, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("storage.go.go NewStorage, pgxpool.ParseConfig(...): %w", err)
	}

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("storage.go.go NewStorage, pgxpool.NewWithConfig(...): %w", err)
	}

	return &Storage{
		db:  db,
		dsn: dsn,
	}, nil
}

func (s *Storage) Migrate(logger *logrus.Logger) error {
	files, _ := migrationFiles.ReadDir("migrations")
	for _, file := range files {
		logger.Infof("Found migration file: %s", file.Name())
	}

	conn, err := sql.Open("pgx", s.dsn)
	if err != nil {
		return fmt.Errorf("storage.go.go Migrate, sql.Open(...) failed: %w", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			logrus.Warnf("storage.go Migrate conn.Close(): %v", err)
		}
	}()

	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationFiles,
		Root:       "migrations",
	}

	n, err := migrate.Exec(conn, "postgres", migrations, migrate.Up)
	if err != nil {
		return fmt.Errorf("storage.go.go Migrate(...) migrate.Exec(...): %w", err)
	}

	logger.Infof("Applied %d migrations successfully", n)

	return nil
}
