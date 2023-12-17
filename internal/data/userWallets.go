package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type UserWalletModel struct {
	DB *sql.DB
}

var (
	ErrDuplicateUserID = errors.New("duplicate user_id")
)

type UserWallet struct {
	ID       int64     `json:"id"`
	UserID   int64     `json:"user_id"`
	Balance  float64   `json:"blance"`
	UpdateAt time.Time `json:"updated_at"`
	Version  int       `json:"-"`
}

func (m UserWalletModel) New(userID int64) error {
	UserWallet := UserWallet{
		UserID:  userID,
		Balance: 10000, // for test purpose
	}
	return m.Insert(UserWallet)
}

func (m UserWalletModel) Insert(wallet UserWallet) error {
	query := `INSERT INTO user_wallets (user_id, blance))
						VALUES($1, $2)`

	args := []any{
		wallet.UserID,
		wallet.Balance,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "user_wallets_user_id_key"`:
			return ErrDuplicateUserID
		default:
			return err
		}
	}

	return nil
}

func (m UserWalletModel) GetUserWallet(userID int64) (*UserWallet, error) {
	query := `SELECT id, user_id, balance, version FROM user_wallets WHERE user_id = $1`
	args := []any{userID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var wallet *UserWallet
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return wallet, nil
}

func (m UserWalletModel) Update(wallet *UserWallet) error {
	query := `UPDATE user_wallets
						SET blance=$1, update_at=NOW(), version = version + 1
						WHERE id=$2 AND version=$3
						RETURNING version`

	args := []any{
		wallet.Balance,
		wallet.ID,
		wallet.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&wallet.Version)
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
