package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/maxwellkuo47/tradingEngine/internal/validator"
)

type OrderModel struct {
	DB DBTX
}

const (
	ORDER_TYPE_BUY = iota
	ORDER_TYPE_SELL
)

const (
	ORDER_PRCIE_TYPE_MARKET = iota
	ORDER_PRICE_TYPE_LIMIT
)

const (
	ORDER_STATUS_KILLED = iota - 1
	ORDER_STATUS_PENDING
	ORDER_STATUS_FILLED
)

var (
	permittedTypeVal      = []int{0, 1}     // 0: buy 1: sell
	permittedPriceTypeVal = []int{0, 1}     // 0: market 1: limit
	permittedStatusVal    = []int{-1, 0, 1} // -1: killed 0: pending 1: filled

)

type Order struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    int64     `json:"user_id"`
	StockID   int64     `json:"stock_id"`
	Type      int       `json:"type"`
	Quantity  int       `json:"quantity"`
	PriceType int       `json:"price_type"`
	Price     float64   `json:"price"`
	Status    int       `json:"status"`
	Version   int       `json:"_"`
}

func ValidateOrder(v *validator.Validator, order Order) {
	v.Check(validator.PermittedValue(order.Type, permittedTypeVal...), "type", "invalid type value")
	v.Check(validator.PermittedValue(order.PriceType, permittedPriceTypeVal...), "price_type", "invalid price_type value")
	v.Check(validator.PermittedValue(order.Status, permittedStatusVal...), "status", "invalid status value")
	v.Check(order.Quantity > 0, "quantity", "quantity must be positive")
	v.Check(order.Price > 0, "price", "price must be positive")
}

func (m OrderModel) Insert(order *Order) error {
	query := `INSERT INTO orders (user_id, stock_id, type, quantity, price_type, price, status)
						VALUES($1, $2, $3, $4, $5, $6, $7)
						RETURNING id, created_at, version`

	args := []any{
		order.UserID,
		order.StockID,
		order.Type,
		order.Quantity,
		order.PriceType,
		order.Price,
		order.Status,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&order.ID,
		&order.CreatedAt,
		&order.Version,
	)

	return err
}
func (m OrderModel) GetOrderForUpdate(orderID int64) (*Order, error) {
	query := `SELECT id, user_id, quantity, price, stock_id, status, version FROM orders
						WHERE id = $1`

	args := []any{orderID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var order Order

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&order.ID,
		&order.UserID,
		&order.Quantity,
		&order.Price,
		&order.StockID,
		&order.Status,
		&order.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &order, nil
}
func (m OrderModel) UpdateOrderStatus(order *Order, staus int) error {
	query := `UPDATE orders SET status = $1, updated_at = $2, version = version + 1
						WHERE id=$3 AND version=$4`

	args := []any{
		staus,
		order.UpdatedAt,
		order.ID,
		order.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
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
