package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Storage struct {
	log *logrus.Logger
	db  *pgxpool.Pool
	dsn string
}

type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (s *Storage) DB() *pgxpool.Pool {
	return s.db
}

func NewStorage(ctx context.Context, log *logrus.Logger, dsn string) (*Storage, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("storage.go.go NewStorage, pgxpool.ParseConfig(...): %w", err)
	}

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("storage.go.go NewStorage, pgxpool.NewWithConfig(...): %w", err)
	}

	return &Storage{
		log: log,
		db:  db,
		dsn: dsn,
	}, nil
}

func (s *Storage) Migrate() error {
	files, _ := migrationFiles.ReadDir("migrations")
	for _, file := range files {
		s.log.Infof("Found migration file: %s", file.Name())
	}

	conn, err := sql.Open("pgx", s.dsn)
	if err != nil {
		return fmt.Errorf("storage.go.go Migrate, sql.Open(...) failed: %w", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			s.log.Warnf("storage.go Migrate conn.Close(): %v", err)
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

	s.log.Infof("Applied %d migrations successfully", n)

	return nil
}
