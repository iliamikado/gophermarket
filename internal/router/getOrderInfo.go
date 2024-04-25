package router

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/iliamikado/gophermarket/internal/config"
	"github.com/iliamikado/gophermarket/internal/db"
	"github.com/iliamikado/gophermarket/internal/logger"
	"github.com/iliamikado/gophermarket/internal/models"
)

const RepeatRequestTime = time.Second * 3
const RepeatRequestTimeOn429 = time.Minute

func updateOrderInfo(order models.Order) {
	resp, err := http.Get(config.AccrualSystemAddress + order.Number)
	if resp != nil {
		defer resp.Body.Close()
	}
	logger.Log("Get info for order " + order.Number)
	
	var newOrder models.Order
	if err != nil || resp.StatusCode != 200 {
		logger.Log(err)
		logger.Log(resp.Status)
		newOrder = order
	} else {
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &newOrder)
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
