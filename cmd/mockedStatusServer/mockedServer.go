package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/iliamikado/gophermarket/internal/models"
)

func main() {
	r := chi.NewRouter()
	r.Get("/api/orders/{number}", getStatus)
	http.ListenAndServe(":8081", r)
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")
	order := models.Order{Number: number, Status: "PROCESSED", Accrual: float64(455.34)}
	body, _ := json.Marshal(order)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}