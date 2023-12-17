package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/maxwellkuo47/tradingEngine/internal/data"
	"github.com/maxwellkuo47/tradingEngine/internal/validator"
)

func (app *application) orderCreate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		StockID   int64   `json:"stock_id"`
		Type      int     `json:"type"`
		Quantity  int     `json:"quantity"`
		PriceType int     `json:"price_type"`
		Price     float64 `json:"price"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badReqResp(w, r, err)
		return
	}
	order := data.Order{
		StockID:   input.StockID,
		Type:      input.Type,
		Quantity:  input.Quantity,
		PriceType: input.PriceType,
		Price:     input.Price,
		Status:    data.ORDER_STATUS_PENDING,
	}
	// validate input data
	v := validator.New()
	if data.ValidateOrder(v, order); !v.Valid() {
		app.failedValidationResp(w, r, v.Errors)
		return
	}

	if exist, err := app.models.Stock.ConfirmStockExist(order.StockID); err != nil || !exist {
		switch {
		case errors.Is(err, data.ErrRecordNotFound) || !exist:
			v.AddError("stock", fmt.Sprintf("can not find stock with id %d", order.StockID))
			app.failedValidationResp(w, r, v.Errors)
		default:
			app.serverErrResp(w, r, err)
		}
		return
	}

	// get user data
	user := app.contextGetUser(r)

	//check and update wallet/stock Balance
	switch order.Type {
	case data.ORDER_TYPE_BUY: // BUY
		wallet, err := app.models.UserWallet.GetUserWallet(user.ID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.balanceRecordNotFoundResp(w, r)
			default:
				app.serverErrResp(w, r, err)
			}
			return
		}
		if order.Price*float64(order.Quantity) > wallet.Balance {
			app.insufficientBalanceResp(w, r)
			return
		}

		// update wallet
		wallet.Balance -= order.Price * float64(order.Quantity)
		err = app.models.UserWallet.Update(wallet)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrEditConflict):
				app.editConflictResp(w, r)
			default:
				app.serverErrResp(w, r, err)
			}
			return
		}
	case data.ORDER_TYPE_SELL: // SELL
		stockBalance, err := app.models.UserStockBalance.GetUserStockBalance(user.ID, order.StockID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.balanceRecordNotFoundResp(w, r)
			default:
				app.serverErrResp(w, r, err)
			}
			return
		}
		if stockBalance.Quantity < order.Quantity {
			app.insufficientBalanceResp(w, r)
			return
		}

		// update stock balance
		stockBalance.Quantity -= order.Quantity
		err = app.models.UserStockBalance.Update(stockBalance)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrEditConflict):
				app.editConflictResp(w, r)
			default:
				app.serverErrResp(w, r, err)
			}
			return
		}
	default:
		panic("invalid type should be eliminate at validate state")
	}
	order.UserID = user.ID

	err = app.models.Order.Insert(order)
	if err != nil {
		app.serverErrResp(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"message": "order create successfully"}, nil)
	if err != nil {
		app.serverErrResp(w, r, err)
		return
	}
}
