package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Stock struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type StockModel struct {
	DB DBTX
}

func (m StockModel) ConfirmStockExist(stock_id int64) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM stocks WHERE id = $1);`
	args := []any{stock_id}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var exist bool

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&exist)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return false, ErrRecordNotFound
		default:
			return false, err
		}
	}
	return exist, err
}

func (m StockModel) GetAllStockIDs() ([]int64, error) {
	query := `SELECT id FROM stocks;`
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var stockIDs []int64

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		err = rows.Scan(
			&id,
		)
		if err != nil {
			return nil, err
		}
		stockIDs = append(stockIDs, id)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return stockIDs, err
}
