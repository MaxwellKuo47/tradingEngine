package data

import (
	"database/sql"
	"errors"
)

type Models struct {
	Users UserModel
	Token TokenModel
}

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

func NewModels(db *sql.DB) Models {
	return Models{
		Users: UserModel{DB: db},
		Token: TokenModel{DB: db},
	}
}
