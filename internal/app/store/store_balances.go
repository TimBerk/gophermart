package store

import (
	"TimBerk/gophermart/internal/app/models/balance"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

func (s *PostgresStore) GetBalance(ctx context.Context, userID int64) (balance.Balance, error) {
	var record balance.Balance

	query := `SELECT current, withdrawn FROM balance WHERE user_id = $1`
	err := s.db.QueryRow(ctx, query, userID).Scan(&record.Current, &record.Withdrawn)
	if err != nil {
		logrus.WithFields(logrus.Fields{"action": "DB.GetBalance", "user": userID, "error": err}).Error("failed to find")
	}
	return record, err
}

func (s *PostgresStore) WithdrawBalance(ctx context.Context, tx pgx.Tx, userID int64, sum float64) error {
	_, err := tx.Exec(ctx, `SELECT 1 FROM balance WHERE user_id = $1 FOR UPDATE`, userID)
	if err != nil {
		return err
	}

	query := `UPDATE balance SET current = current - $2, withdrawn = withdrawn + $2 WHERE user_id = $1`
	_, err = tx.Exec(ctx, query, userID, sum)
	return err
}

func (s *PostgresStore) AddBalance(ctx context.Context, tx pgx.Tx, userID int64, sum float64) error {
	query := `UPDATE balance SET current = current + $2 WHERE user_id = $1`
	_, err := tx.Exec(ctx, query, userID, sum)
	return err
}

func (s *PostgresStore) GetOrderWithdrawals(ctx context.Context, userID int64) (balance.WithdrawnList, error) {
	query := `SELECT order_number, sum, created_at FROM withdrawals WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records balance.WithdrawnList
	for rows.Next() {
		var record balance.WithdrawnResponse
		if errRow := rows.Scan(&record.Number, &record.Sum, &record.CreatedAt); errRow != nil {
			logrus.WithFields(logrus.Fields{"action": "DB.GetOrderWithdrawals", "user": userID, "error": errRow}).Error("failed to find order")
			return nil, errRow
		}
		records = append(records, record)
	}

	if errRow := rows.Err(); errRow != nil {
		logrus.WithFields(logrus.Fields{"action": "DB.GetOrderWithdrawals", "user": userID, "error": errRow}).Error("failed to find orders")
		return nil, errRow
	}

	return records, err
}
