package store

import (
	model "TimBerk/gophermart/internal/app/models/order"
	"context"
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
)

type Status string

type OrderRecord struct {
	ID      int64
	UserID  int64
	Order   string
	Status  Status
	Accrual sql.NullFloat64
}

const (
	New        Status = "NEW"
	Processing Status = "PROCESSING"
	Invalid    Status = "INVALID"
	Processed  Status = "PROCESSED"
	Undefined  Status = "UNDEFINED"
)

func GetConstStatus(status string) Status {
	statusMap := map[string]Status{
		"NEW":        New,
		"PROCESSING": Processing,
		"INVALID":    Invalid,
		"PROCESSED":  Processed,
		"UNDEFINED":  Undefined,
	}
	mainStatus, exists := statusMap[status]
	if exists {
		return mainStatus
	}
	return Undefined
}

func (s *PostgresStore) AddOrder(ctx context.Context, userID int64, order string) error {
	query := `INSERT INTO orders (user_id, order_number) VALUES ($1, $2)`
	_, err := s.db.Exec(ctx, query, userID, order)
	return err
}

func (s *PostgresStore) GetOrder(ctx context.Context, order string) (OrderRecord, error) {
	var record OrderRecord
	query := `SELECT id, user_id, order_number, status, accrual FROM orders WHERE order_number = $1`
	err := s.db.QueryRow(ctx, query, order).Scan(&record.ID, &record.UserID, &record.Order, &record.Status, &record.Accrual)
	return record, err
}

func (s *PostgresStore) GetOrderList(ctx context.Context, userID int64) (model.OrderListResponse, error) {
	query := `SELECT order_number, status, accrual, created_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records model.OrderListResponse
	for rows.Next() {
		var record model.OrderResponse
		if errRows := rows.Scan(&record.Number, &record.Status, &record.Accrual, &record.CreatedAt); errRows != nil {
			logrus.WithFields(logrus.Fields{"action": "DB.GetOrderList", "user": userID, "error": errRows}).Error("failed to find order")
			return nil, errRows
		}
		records = append(records, record)
	}
	if errRows := rows.Err(); errRows != nil {
		logrus.WithFields(logrus.Fields{"action": "DB.GetOrderList", "user": userID, "error": errRows}).Error("failed to find orders")
		return nil, errRows
	}
	return records, err
}

func (s *PostgresStore) AddWithdrawal(ctx context.Context, userID int64, order string, sum float64) error {
	tx, err := s.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin transcation: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO withdrawals (user_id, order_number, sum) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, query, userID, order, sum)
	if err != nil {
		return fmt.Errorf("create withdrawals error: %w", err)
	}

	err = s.WithdrawBalance(ctx, tx, userID, sum)
	if err != nil {
		return fmt.Errorf("update user balance error: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return err
}

func (s *PostgresStore) GetOrdersForAccrual(ctx context.Context) ([]model.UserOrder, error) {
	query := `SELECT user_id, order_number, status FROM orders WHERE status in ('NEW', 'PROCESSING')`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.UserOrder
	for rows.Next() {
		var o model.UserOrder
		if errRows := rows.Scan(&o.UserID, &o.Number, &o.Status); errRows != nil {
			logrus.WithFields(logrus.Fields{"action": "DB.GetOrdersForAccrual", "error": errRows}).Error("failed to find order")
			return nil, errRows
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (s *PostgresStore) UpdateOrderStatus(ctx context.Context, userID int64, order string, status Status, accrual float64) error {
	tx, err := s.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin transcation: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `UPDATE orders SET status = $1, accrual = $2 WHERE order_number = $3`
	_, err = s.db.Exec(ctx, query, status, accrual, order)
	if err != nil {
		return fmt.Errorf("update order status with accrual error: %w", err)
	}

	err = s.AddBalance(ctx, tx, userID, accrual)
	if err != nil {
		return fmt.Errorf("update user balance error: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return err
}
