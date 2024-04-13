package router

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/iliamikado/gophermarket/internal/config"
	"github.com/iliamikado/gophermarket/internal/db"
	"github.com/iliamikado/gophermarket/internal/logger"
	"github.com/iliamikado/gophermarket/internal/models"
)

const MaxResponseTime = time.Second * 3

func getOrderInfo(orderNumber string) models.Order {
	client := &http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), MaxResponseTime)
	defer cancel()
	req, err1 := http.NewRequestWithContext(ctx, "GET", config.AccrualSystemAddress + orderNumber, nil)
	resp, err2 := client.Do(req)
	logger.Log("Get info for order " + orderNumber)
	if err1 != nil || err2 != nil || resp.StatusCode != 200 {
		logger.Log(err1.Error())
		logger.Log(err2.Error())
		logger.Log(resp.Status)
		return models.Order{Number: orderNumber}
	}
	var order models.Order
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	dec.Decode(&order)
	logger.Log("Finally get " + orderNumber + " " + order.Status)
	go db.UpdateOrder(order)
	return order
}