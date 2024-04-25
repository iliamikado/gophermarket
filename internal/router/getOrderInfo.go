package router

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/iliamikado/gophermarket/internal/config"
	"github.com/iliamikado/gophermarket/internal/db"
	"github.com/iliamikado/gophermarket/internal/logger"
	"github.com/iliamikado/gophermarket/internal/models"
)

const MaxResponseTime = time.Second * 3
const RepeatRequestTime = time.Second * 3
const RepeatRequestTimeOn429 = time.Minute

func updateOrderInfo(order models.Order) {
	logger.Log("Get info for order " + order.Number)
	
	newOrder, needWait := getOrderInfo(order)

	if newOrder.Status != "PROCESSED" && newOrder.Status != "INVALID" {
		go func(order models.Order) {
			var sleep = RepeatRequestTime
			if needWait {
				sleep = RepeatRequestTimeOn429
			}
			time.Sleep(sleep)
			updateOrderInfo(order)
		}(newOrder)
	}

	logger.Log("Finally get " + newOrder.Number + ":")
	logger.Log(newOrder)
	if order != newOrder {
		db.UpdateOrder(newOrder)
	}
}

func getOrderInfo(order models.Order) (models.Order, bool) {
	client := &http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), MaxResponseTime)
	defer cancel()
	req, err1 := http.NewRequestWithContext(ctx, "GET", config.AccrualSystemAddress + order.Number, nil)
	if err1 != nil {
		return order, false
	}
	resp, err2 := client.Do(req)
	if err2 != nil {
		return order, false
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return order, false
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return order, true
	}

	var newOrder models.Order
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &newOrder)
	newOrder.Number = order.Number
	return newOrder, false
}
