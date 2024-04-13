package router

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/iliamikado/gophermarket/internal/config"
	"github.com/iliamikado/gophermarket/internal/models"
)

const MaxResponseTime = time.Second * 3

func getOrderInfo(orderNumber string) models.Order {
	client := &http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), MaxResponseTime)
	defer cancel()
	req, err1 := http.NewRequestWithContext(ctx, "GET", config.AccrualSystemAddress+ "/" + orderNumber, nil)
	resp, err2 := client.Do(req)
	if err1 != nil || err2 != nil {
		return models.Order{Number: orderNumber}
	}
	var order models.Order
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	dec.Decode(&order)
	return order
}