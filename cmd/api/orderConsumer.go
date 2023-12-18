package main

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/maxwellkuo47/tradingEngine/internal/data"
)

func (app *application) spinUpConsumer() error {
	stockIDs, err := app.models.Stock.GetAllStockIDs()
	if err != nil {
		return err
	}
	for _, stockID := range stockIDs {
		app.createBuyOrderConsumer(stockID)
		app.createSellOrderConsumer(stockID)
	}
	return nil
}
func (app *application) createBuyOrderConsumer(stockID int64) {
	goroutineName := fmt.Sprintf("buyOrderConsumer_%d", stockID)
	app.background(goroutineName, func() {
		for {
			select {
			case <-app.done: // for gracefully shutdown
				app.infoLogger.Info("stop buyOrderConsumer", slog.Int64("stock_id", stockID))
			default:
				price, _ := app.mockStockPrices.Load(stockID)
				currentPrice := price.(float64)
				currentTime := time.Now()
				redisOrder, err := app.consumeBuyOrder(stockID, currentPrice)
				if err != nil {
					app.errorLogger.Error("error consumeBuyOrder", slog.Int64("consumer_stock_id", stockID), slog.String("msg", err.Error()), slog.String("state", "get order from queue"))
				} else if redisOrder != nil {
					// pretend we're successfully sending the sell request to the platform for the customer then we need to do
					// 1. update order status
					// 2. update user wallet there is difference between actual price and order price
					// 3. update user stock balance
					// 4. create trade record
					orderProcessName := fmt.Sprintf("stock_%d_process_buy_order_%d", stockID, redisOrder.OrderID)
					// create a goroutine to process db transaction
					// consumer just go for next order
					app.background(orderProcessName, func() {
						app.processBuyOrder(stockID, redisOrder.OrderID, redisOrder.UserID, currentPrice, currentTime)
					})
				}
				time.Sleep(time.Millisecond * 50)
			}
		}
	})
}

func (app *application) createSellOrderConsumer(stockID int64) {
	goroutineName := fmt.Sprintf("sellOrderConsumer_%d", stockID)
	app.background(goroutineName, func() {
		for {
			select {
			case <-app.done: // for gracefully shutdown
				app.infoLogger.Info("stop sellOrderConsumer", slog.Int64("stock_id", stockID))
			default:
				price, _ := app.mockStockPrices.Load(stockID)
				currentPrice := price.(float64)
				currentTime := time.Now()
				redisOrder, err := app.consumeSellOrder(stockID, currentPrice)
				if err != nil {
					app.errorLogger.Error("error consumeSellOrder", slog.Int64("consumer_stock_id", stockID), slog.String("msg", err.Error()), slog.String("state", "get order from queue"))
				} else if redisOrder != nil {
					// pretend we're successfully sending the sell request to the platform for the customer then we need to do
					// 1. update order status
					// 2. update user wallet with the actual sell price
					// 3. create trade record
					orderProcessName := fmt.Sprintf("stock_%d_process_sell_order_%d", stockID, redisOrder.OrderID)
					app.background(orderProcessName, func() {
						app.processSellOrder(stockID, redisOrder.OrderID, redisOrder.UserID, currentPrice, currentTime)
					})
				}
				time.Sleep(time.Millisecond * 50)
			}
		}
	})
}

func (app *application) processBuyOrder(stockID, orderID, userID int64, currentPrice float64, currentTime time.Time) {
	// begin transaction
	tx, err := app.models.DBHandler.Begin()
	defer tx.Rollback()
	if err != nil {
		app.errorLogger.Error(
			"error Begin",
			slog.Int64("consumer_stock_id", stockID),
			slog.Int64("order_id", orderID),
			slog.String("msg", err.Error()),
			slog.String("state", "begin transaction"),
		)
	}
	txModels := data.NewTxModels(tx)
	// get order and update status
	order, err := txModels.Order.GetOrderForUpdate(orderID)
	if err != nil {
		app.errorLogger.Error(
			"error GetOrderForUpdate",
			slog.Int64("consumer_stock_id", stockID),
			slog.Int64("order_id", orderID),
			slog.String("msg", err.Error()),
			slog.String("state", "get order record"),
		)
		return
	}
	order.UpdatedAt = currentTime
	txModels.Order.UpdateOrderStatus(order, data.ORDER_STATUS_FILLED)

	// update user's wallet if actual price is lower than order price
	if currentPrice < order.Price {
		userWallet, err := txModels.UserWallet.GetUserWallet(userID)
		if err != nil {
			app.errorLogger.Error(
				"error GetUserWallet",
				slog.Int64("consumer_stock_id", stockID),
				slog.Int64("order_id", orderID),
				slog.String("msg", err.Error()),
				slog.String("state", "get user wallet"),
			)
			return
		}
		// refund difference to user
		userWallet.Balance += float64(order.Quantity) * (order.Price - currentPrice)
		err = txModels.UserWallet.Update(userWallet)
		if err != nil {
			app.errorLogger.Error(
				"error Update",
				slog.Int64("consumer_stock_id", stockID),
				slog.Int64("order_id", orderID),
				slog.String("msg", err.Error()),
				slog.String("state", "update user wallet"),
			)
			return
		}
	}

	// update user's stock balance
	// get balance
	stockBalance, err := txModels.UserStockBalance.GetUserStockBalance(userID, stockID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			// no stock balance record just create one
			stockBalance = &data.UserStockBalance{
				UserID:   userID,
				StockID:  stockID,
				Quantity: 0,
			}
			err = txModels.UserStockBalance.Insert(stockBalance)
			if err != nil {
				app.errorLogger.Error(
					"error Insert",
					slog.Int64("consumer_stock_id", stockID),
					slog.Int64("order_id", orderID),
					slog.String("msg", err.Error()),
					slog.String("state", "create user stock balance record"),
				)
				return
			}
		default:
			app.errorLogger.Error(
				"error GetUserStockBalance",
				slog.Int64("consumer_stock_id", stockID),
				slog.Int64("order_id", orderID),
				slog.String("msg", err.Error()),
				slog.String("state", "get user stock balance record"),
			)
			return
		}
	}

	// update stock balance
	stockBalance.Quantity += order.Quantity
	err = txModels.UserStockBalance.Update(stockBalance)
	if err != nil {
		app.errorLogger.Error(
			"error Update",
			slog.Int64("consumer_stock_id", stockID),
			slog.Int64("order_id", orderID),
			slog.String("msg", err.Error()),
			slog.String("state", "update user stock balance record"),
		)
		return
	}

	// create trade record
	trade := data.Trade{
		UserID:     order.UserID,
		OrderID:    order.ID,
		Quantity:   order.Quantity,
		Price:      currentPrice,
		ExecutedAt: currentTime,
	}

	err = txModels.Trade.Insert(trade)
	if err != nil {
		app.errorLogger.Error(
			"error Insert",
			slog.Int64("consumer_stock_id", stockID),
			slog.Int64("order_id", orderID),
			slog.String("msg", err.Error()),
			slog.String("state", "insert trade record"),
		)
		return
	}

	tx.Commit()
}

func (app *application) processSellOrder(stockID, orderID, userID int64, currentPrice float64, currentTime time.Time) {
	// begin transaction
	tx, err := app.models.DBHandler.Begin()
	defer tx.Rollback()
	if err != nil {
		app.errorLogger.Error(
			"error Begin",
			slog.Int64("consumer_stock_id", stockID),
			slog.Int64("order_id", orderID),
			slog.String("msg", err.Error()),
			slog.String("state", "begin transaction"),
		)
	}
	txModels := data.NewTxModels(tx)
	// get order and update status
	order, err := txModels.Order.GetOrderForUpdate(orderID)
	if err != nil {
		app.errorLogger.Error(
			"error GetOrderForUpdate",
			slog.Int64("consumer_stock_id", stockID),
			slog.Int64("order_id", orderID),
			slog.String("msg", err.Error()),
			slog.String("state", "get order record"),
		)
		return
	}
	order.UpdatedAt = currentTime
	txModels.Order.UpdateOrderStatus(order, data.ORDER_STATUS_FILLED)

	// update user's wallet
	userWallet, err := txModels.UserWallet.GetUserWallet(userID)
	if err != nil {
		app.errorLogger.Error(
			"error GetUserWallet",
			slog.Int64("consumer_stock_id", stockID),
			slog.Int64("order_id", orderID),
			slog.String("msg", err.Error()),
			slog.String("state", "get user wallet"),
		)
		return
	}

	// because currentPrice may higher than order price so using currentPrice to calculate
	userWallet.Balance += float64(order.Quantity) * currentPrice
	err = txModels.UserWallet.Update(userWallet)
	if err != nil {
		app.errorLogger.Error(
			"error Update",
			slog.Int64("consumer_stock_id", stockID),
			slog.Int64("order_id", orderID),
			slog.String("msg", err.Error()),
			slog.String("state", "update user wallet"),
		)
		return
	}

	// no need to update stock balance because it was handled in createOrderHandler

	// create trade record
	trade := data.Trade{
		UserID:     order.UserID,
		OrderID:    order.ID,
		Quantity:   order.Quantity,
		Price:      currentPrice,
		ExecutedAt: currentTime,
	}

	err = txModels.Trade.Insert(trade)
	if err != nil {
		app.errorLogger.Error(
			"error Insert",
			slog.Int64("consumer_stock_id", stockID),
			slog.Int64("order_id", orderID),
			slog.String("msg", err.Error()),
			slog.String("state", "insert trade record"),
		)
		return
	}

	tx.Commit()
}
