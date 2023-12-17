package main

import (
	"net/http"
)

func (app *application) adjustStockPrice(w http.ResponseWriter, r *http.Request) {
	var input struct {
		StockID int64   `json:"stock_id"`
		Price   float64 `json:"price"`
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
