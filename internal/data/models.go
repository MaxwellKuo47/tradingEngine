package data

import (
	"context"
	"database/sql"
	"errors"
)

type DBModels struct {
	DBHandler        *sql.DB
	Users            UserModel
	Token            TokenModel
	Order            OrderModel
	Trade            TradeModel
	Stock            StockModel
	UserWallet       UserWalletModel
	UserStockBalance UserStockBalanceModel
}
type TxModels struct {
	Users            UserModel
	Token            TokenModel
	Order            OrderModel
	Trade            TradeModel
	Stock            StockModel
	UserWallet       UserWalletModel
	UserStockBalance UserStockBalanceModel
}

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type DBTX interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func NewModels(db *sql.DB) DBModels {
	return DBModels{
		DBHandler:        db,
		Users:            UserModel{DB: db},
		Token:            TokenModel{DB: db},
		Order:            OrderModel{DB: db},
		Trade:            TradeModel{DB: db},
		Stock:            StockModel{DB: db},
		UserWallet:       UserWalletModel{DB: db},
		UserStockBalance: UserStockBalanceModel{DB: db},
	}
}

func NewTxModels(tx *sql.Tx) TxModels {
	return TxModels{
		Users:            UserModel{DB: tx},
		Token:            TokenModel{DB: tx},
		Order:            OrderModel{DB: tx},
		Trade:            TradeModel{DB: tx},
		Stock:            StockModel{DB: tx},
		UserWallet:       UserWalletModel{DB: tx},
		UserStockBalance: UserStockBalanceModel{DB: tx},
	}
}
