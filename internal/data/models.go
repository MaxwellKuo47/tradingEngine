package data

import (
	"database/sql"
	"errors"
)

type Models struct {
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

func NewModels(db *sql.DB) Models {
	return Models{
		Users:            UserModel{DB: db},
		Token:            TokenModel{DB: db},
		Order:            OrderModel{DB: db},
		Trade:            TradeModel{DB: db},
		Stock:            StockModel{DB: db},
		UserWallet:       UserWalletModel{DB: db},
		UserStockBalance: UserStockBalanceModel{DB: db},
	}
}
