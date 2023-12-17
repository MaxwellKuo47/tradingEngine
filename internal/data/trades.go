package data

import (
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
