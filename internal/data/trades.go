package data

import (
	"database/sql"
	"time"
)

type TradeModel struct {
	DB *sql.DB
}

type Trade struct {
	ID          int64     `json:"id"`
	ExecutedAt  time.Time `json:"executed_at"`
	BuyOrderID  int64     `json:"buy_order_id"`
	SellOrderID int64     `json:"sell_order_id"`
	Quantity    int       `json:"quantity"`
	Price       float64   `json:"price"`
}
