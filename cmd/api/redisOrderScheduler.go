package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/maxwellkuo47/tradingEngine/internal/data"
	"github.com/redis/go-redis/v9"
)

func createRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return client, nil
}

type storedOrder struct {
	OrderID    int64     `json:"order_id"`
	UserID     int64     `json:"user_id"`
	StockID    int64     `json:"stock_id"`
	Quantity   int       `json:"quantity"`
	CreateTime time.Time `json:"create_time"` // just for demo purpose
}

func (app *application) insertBuyOrder(order data.Order) error {
	buyHeapKey := fmt.Sprintf("buy_heap_%d", order.StockID)
	buyQueueKey := fmt.Sprintf("buy_%d_at_%f", order.StockID, order.Price)
	member := []redis.Z{
		{
			Score:  order.Price,
			Member: buyQueueKey,
		},
	}
	if err := app.redisClient.ZAdd(context.Background(), buyHeapKey, member...).Err(); err != nil {
		return err
	}

	storedOrder := storedOrder{
		OrderID:    order.ID,
		UserID:     order.UserID,
		StockID:    order.StockID,
		Quantity:   order.Quantity,
		CreateTime: time.Now(),
	}
	storedOrderJSON, err := json.Marshal(storedOrder)
	if err != nil {
		return err
	}

	if err := app.redisClient.RPush(context.Background(), buyQueueKey, storedOrderJSON).Err(); err != nil {
		return err
	}

	return nil
}

func (app *application) insertSellOrder(order data.Order) error {
	sellHeapKey := fmt.Sprintf("sell_heap_%d", order.StockID)
	sellQueueKey := fmt.Sprintf("sell_%d_at_%f", order.StockID, order.Price)
	member := []redis.Z{
		{
			Score:  order.Price,
			Member: sellQueueKey,
		},
	}
	if err := app.redisClient.ZAdd(context.Background(), sellHeapKey, member...).Err(); err != nil {
		return err
	}

	storedOrder := storedOrder{
		OrderID:    order.ID,
		UserID:     order.UserID,
		StockID:    order.StockID,
		Quantity:   order.Quantity,
		CreateTime: time.Now(),
	}

	storedOrderJSON, err := json.Marshal(storedOrder)
	if err != nil {
		return err
	}

	if err := app.redisClient.RPush(context.Background(), sellQueueKey, storedOrderJSON).Err(); err != nil {
		return err
	}

	return nil
}

func (app *application) consumeBuyOrder(stockID int64, currentPrice float64) (*storedOrder, error) {
	// Key for the heap
	buyHeapKey := fmt.Sprintf("buy_heap_%d", stockID)

	// Get the highest price from the heap
	highest, err := app.redisClient.ZRevRangeWithScores(context.Background(), buyHeapKey, 0, 0).Result()
	if err != nil {
		return nil, err
	}
	if len(highest) == 0 {
		return nil, nil
	}

	// Construct the queue key using the highest price
	highestPrice := highest[0].Score
	if highestPrice < currentPrice {
		return nil, nil
	}

	buyQueueKey := fmt.Sprintf("buy_%d_at_%f", stockID, highestPrice)

	// Left pop the order from the corresponding queue
	orderJSON, err := app.redisClient.LPop(context.Background(), buyQueueKey).Result()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			// Check if the list is empty
			app.redisClient.Del(context.Background(), buyQueueKey)
			app.redisClient.ZRem(context.Background(), buyHeapKey, buyQueueKey)
			return nil, nil
		default:
			return nil, err
		}
	}

	// Unmarshal the order
	var order storedOrder
	err = json.Unmarshal([]byte(orderJSON), &order)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (app *application) consumeSellOrder(stockID int64, currentPrice float64) (*storedOrder, error) {
	// Key for the heap
	sellHeapKey := fmt.Sprintf("sell_heap_%d", stockID)

	// Get the lowest price from the heap
	lowest, err := app.redisClient.ZRangeWithScores(context.Background(), sellHeapKey, 0, 0).Result()
	if err != nil {
		return nil, err
	}
	if len(lowest) == 0 {
		return nil, nil
	}

	// Construct the queue key using the lowest price
	lowestPrice := lowest[0].Score
	if lowestPrice > currentPrice {
		return nil, nil
	}

	sellQueueKey := fmt.Sprintf("sell_%d_at_%f", stockID, lowestPrice)

	// Left pop the order from the corresponding queue
	orderJSON, err := app.redisClient.LPop(context.Background(), sellQueueKey).Result()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			// Check if the list is empty
			app.redisClient.Del(context.Background(), sellQueueKey)
			app.redisClient.ZRem(context.Background(), sellHeapKey, sellQueueKey)
			return nil, nil
		default:
			return nil, err
		}
	}

	// Unmarshal the order
	var order storedOrder
	err = json.Unmarshal([]byte(orderJSON), &order)
	if err != nil {
		return nil, err
	}

	return &order, nil
}
