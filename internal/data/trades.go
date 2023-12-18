package data

import (
	"context"
	"time"
)

type TradeModel struct {
	DB DBTX
}

type Trade struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	OrderID    int64     `json:"order_id"`
	Quantity   int       `json:"quantity"`
	Price      float64   `json:"price"`
	ExecutedAt time.Time `json:"executed_at"`
}

func (m TradeModel) Insert(trade Trade) error {

	query := `INSERT INTO trades (user_id, order_id, quantity, price, executed_at)`

	args := []any{
		trade.UserID,
		trade.OrderID,
		trade.Quantity,
		trade.Price,
		trade.ExecutedAt,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}
