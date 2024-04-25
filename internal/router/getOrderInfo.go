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
	client := &http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), MaxResponseTime)
	defer cancel()
	req, err1 := http.NewRequestWithContext(ctx, "GET", config.AccrualSystemAddress + order.Number, nil)
	resp, err2 := client.Do(req)
	defer resp.Body.Close()
	logger.Log("Get info for order " + order.Number)
	
	var newOrder models.Order
	if err1 != nil || err2 != nil || resp.StatusCode != 200 {
		logger.Log(resp.Status)
		newOrder = order
	} else {
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &newOrder)
		resp.Body.Close()
		newOrder.Number = order.Number
	}

	if newOrder.Status != "PROCESSED" && newOrder.Status != "INVALID" {
		go func(order models.Order) {
			var sleep = RepeatRequestTime
			if resp.StatusCode == http.StatusTooManyRequests {
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
