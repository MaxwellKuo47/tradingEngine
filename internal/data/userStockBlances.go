package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type UserStockBalanceModel struct {
	DB DBTX
}

type UserStockBalance struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	StockID   int64     `json:"stock_id"`
	Quantity  int       `json:"quantity"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"`
}

func (m UserStockBalanceModel) Insert(stockBalance *UserStockBalance) error {
	query := `INSERT INTO user_stock_balances (user_id, stock_id, quantity, updated_at)
						VALUES ($1, $2, $3, NOW()) 
						RETURNING id, version`

	args := []any{
		stockBalance.UserID,
		stockBalance.StockID,
		stockBalance.Quantity,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&stockBalance.ID,
		&stockBalance.Version,
	)

	if err != nil {
		return err
	}

	return nil
}
func (m UserStockBalanceModel) GetUserStockBalance(userID int64, stockID int64) (*UserStockBalance, error) {
	query := `SELECT id, user_id, stock_id, quantity, version 
						FROM user_stock_balances 
						WHERE user_id = $1 AND stock_id = $2`

	args := []any{
		userID,
		stockID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var stockBalance UserStockBalance
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&stockBalance.ID, &stockBalance.UserID, &stockBalance.StockID, &stockBalance.Quantity, &stockBalance.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &stockBalance, nil
}

func (m UserStockBalanceModel) Update(stockBalance *UserStockBalance) error {
	query := `UPDATE user_stock_balances 
						SET quantity = $1, updated_at = NOW(), version = version + 1 
						WHERE id = $2 AND version = $3
						RETURNING version`
	args := []any{
		stockBalance.Quantity,
		stockBalance.ID,
		stockBalance.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&stockBalance.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}
