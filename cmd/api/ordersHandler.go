package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/maxwellkuo47/tradingEngine/internal/data"
	"github.com/maxwellkuo47/tradingEngine/internal/validator"
)

func (app *application) orderCreateHandler(w http.ResponseWriter, r *http.Request) {
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

	if order.PriceType == data.ORDER_PRCIE_TYPE_MARKET {
		currentStockPrice, _ := app.mockStockPrices.Load(order.StockID)
		// let the order could be consumed immediately
		switch order.Type {
		case data.ORDER_TYPE_BUY:
			order.Price = currentStockPrice.(float64) + 10.0
		case data.ORDER_TYPE_SELL:
			order.Price = currentStockPrice.(float64) - 10.0
		default:
			//just ignore because this request would be return by validator
		}
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
	tx, err := app.models.DBHandler.Begin()
	if err != nil {
		app.serverErrResp(w, r, err)
		return
	}
	defer tx.Rollback()
	txModels := data.NewTxModels(tx)
	//check and update wallet/stock Balance
	switch order.Type {
	case data.ORDER_TYPE_BUY: // BUY
		wallet, err := txModels.UserWallet.GetUserWallet(user.ID)
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
		err = txModels.UserWallet.Update(wallet)
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
		stockBalance, err := txModels.UserStockBalance.GetUserStockBalance(user.ID, order.StockID)
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
		err = txModels.UserStockBalance.Update(stockBalance)
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

	err = txModels.Order.Insert(&order)
	if err != nil {
		app.serverErrResp(w, r, err)
		return
	}

	switch order.Type {
	case data.ORDER_TYPE_BUY:
		err = app.insertBuyOrder(order)
		if err != nil {
			app.serverErrResp(w, r, err)
			return
		}
	case data.ORDER_TYPE_SELL:
		err = app.insertSellOrder(order)
		if err != nil {
			app.serverErrResp(w, r, err)
			return
		}
	default:
		panic("invalid type should be eliminate at insert state")
	}
	tx.Commit()
	err = app.writeJSON(w, http.StatusCreated, envelope{"message": "order create successfully"}, nil)
	if err != nil {
		app.serverErrResp(w, r, err)
		return
	}
}
