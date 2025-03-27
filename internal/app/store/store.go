package store

import (
	"TimBerk/gophermart/internal/app/settings/config"
	"context"
	"github.com/jackc/pgx/v5/stdlib"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
)

type PostgresStore struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewPgPool(ctx context.Context, connString string) (*PostgresStore, error) {
	var pgInstance *PostgresStore
	var pgOnce sync.Once
	var pgErr error

	pgOnce.Do(func() {
		db, err := pgxpool.New(ctx, connString)
		if err != nil {
			pgErr = err
			return
		}

		pgInstance = &PostgresStore{db: db}
	})

	if pgErr != nil {
		logrus.WithField("error", pgErr).Error("unable to create connection pool")
		return nil, pgErr
	}

	return pgInstance, nil
}

func NewPostgresStore(cfg *config.Config) (*PostgresStore, error) {
	ctx := context.Background()

	pgStore, err := NewPgPool(ctx, cfg.DatabaseURI)
	if err != nil {
		return pgStore, err
	}

	pgStore.cfg = cfg

	pgStore.initDB()

	return pgStore, nil
}

func (s *PostgresStore) initDB() {
	migrationDir := "migrations"

	conn, err := s.db.Acquire(context.Background())
	if err != nil {
		logrus.WithField("error", err).Error("failed to acquire connection")
		return
	}
	defer conn.Release()

	if err = goose.SetDialect("postgres"); err != nil {
		logrus.WithField("error", err).Error("failed to set dialect")
		return
	}

	db := stdlib.OpenDBFromPool(s.db)

	if err = goose.Up(db, migrationDir); err != nil {
		logrus.WithField("error", err).Error("failed to run migrations")
		return
	}

	logrus.Info("Migrations applied successfully!")
}

func (s *PostgresStore) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return s.db.Begin(ctx)
}
