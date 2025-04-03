package store

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type UserRecord struct {
	ID           int64
	Username     string
	PasswordHash string
}

func (s *PostgresStore) AddUser(ctx context.Context, username string, password string) (int64, error) {
	var userID int64

	query := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id`
	err := s.db.QueryRow(ctx, query, username, password).Scan(&userID)
	if err != nil {
		logrus.WithFields(logrus.Fields{"action": "DB.AddUser", "user": username, "error": err}).Error("failed to create user")
		return 0, err
	}
	return userID, err
}

func (s *PostgresStore) CheckUser(ctx context.Context, username string) (int64, error) {
	var exists int64
	query := `SELECT id FROM users WHERE username = $1`
	err := s.db.QueryRow(ctx, query, username).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		logrus.WithFields(logrus.Fields{"action": "DB.CheckUser", "user": username, "error": err}).Info("failed to find user")
		return 0, nil
	}
	return exists, err
}

func (s *PostgresStore) GetUser(ctx context.Context, username string) (UserRecord, error) {
	var userRecord UserRecord
	query := `SELECT id, username, password_hash FROM users WHERE username = $1`
	err := s.db.QueryRow(ctx, query, username).Scan(&userRecord.ID, &userRecord.Username, &userRecord.PasswordHash)
	return userRecord, err
}
