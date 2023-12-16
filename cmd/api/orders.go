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

	user := app.contextGetUser(r)
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
