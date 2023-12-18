package main

import (
	"net/http"
	"slices"
)

func (app *application) adjustStockPrice(w http.ResponseWriter, r *http.Request) {
	var input struct {
		StockID int64   `json:"stock_id"`
		Price   float64 `json:"price"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badReqResp(w, r, err)
		return
	}

	stockIDs, err := app.models.Stock.GetAllStockIDs()
	if err != nil {
		app.serverErrResp(w, r, err)
		return
	}

	if exist := slices.Contains(stockIDs, input.StockID); !exist {
		app.failedValidationResp(w, r, map[string]string{"stock_id": "stock id does not exist"})
	}

	if input.Price < 0 {
		app.failedValidationResp(w, r, map[string]string{"price": "price cannot be negative"})
	}

	app.mockStockPrices.Store(input.StockID, input.Price)

	err = app.writeJSON(w, http.StatusAccepted, envelope{"params": input}, nil)
	if err != nil {
		app.serverErrResp(w, r, err)
		return
	}

}

func (app *application) createFakeStockPricesForTesting() error {
	ids, err := app.models.Stock.GetAllStockIDs()

	if err != nil {
		return err
	}

	for _, id := range ids {
		app.mockStockPrices.Store(id, 100)
	}

	return nil
}
