package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/iliamikado/gophermarket/internal/db"
	"github.com/iliamikado/gophermarket/internal/logger"
	"github.com/iliamikado/gophermarket/internal/models"
)

func AppRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/user/register", register)
	r.Post("/api/user/login", login)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/api/user/orders", postOrder)
		r.Get("/api/user/orders", getOrders)
		r.Get("/api/user/balance", getBalance)
		r.Post("/api/user/balance/withdraw", pointsWithdraw)
		r.Get("/api/user/withdrawals", getWithdrawals)
	})
	r.Get("/mock/api/orders/{number}", mockedGetOrderStatus)
	return r
}

func register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := readBody(r, &user)

	logger.Log("Register user " + user.Login)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if db.IsLoginExist(user.Login) {
		w.WriteHeader(http.StatusConflict)
		return
	}
	db.AddNewUser(user.Login, user.Password)
	http.SetCookie(w, &http.Cookie{Name: "JWT", Value: buildJWTString(user.Login), Path: "/"})
	w.WriteHeader(http.StatusOK)
}

func login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := readBody(r, &user)

	logger.Log("Login user " + user.Login)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !db.IsValidUser(user.Login, user.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "JWT", Value: buildJWTString(user.Login), Path: "/"})
	w.WriteHeader(http.StatusOK)
}

func postOrder(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value(userLoginKey{}).(string)
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	number := string(body)
	logger.Log("Post order number " + number + ", login - " + login)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !checkOrderNumber(number) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	ordersUserLogin, exists := db.FindOrder(number)
	if exists {
		if ordersUserLogin == login {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusConflict)
		}
		return
	}
	order := models.Order{Number: number, Status: "NEW"}
	db.AddNewOrder(order, login)
	go func() {
		updateOrderInfo(order)
	}()
	w.WriteHeader(http.StatusAccepted)
}

func getOrders(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value(userLoginKey{}).(string)
	orders := db.GetUsersOrders(login)
	logger.Log("Get orders from login " + login + ":")
	logger.Log(orders)
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	resp, _ := json.Marshal(orders)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func getBalance(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value(userLoginKey{}).(string)
	logger.Log("Getting balance for login - " + login)
	orders := db.GetUsersOrders(login)
	var sum float64 = 0
	for _, order := range orders {
		sum += order.Accrual
	}
	withdrawn := db.GetWithdrawn(login)
	logger.Log(fmt.Sprintf("Sum in orders = %g, withdrawn = %g", sum, withdrawn))
	ans := models.Balance{Current: sum - withdrawn, Withdrawn: withdrawn}
	body, _ := json.Marshal(ans)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func pointsWithdraw(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value(userLoginKey{}).(string)
	logger.Log("Withdraw points from login - " + login)
	var withdrawReq models.WithdrawRequest
	readBody(r, &withdrawReq)

	if !checkOrderNumber(withdrawReq.Order) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	orders := db.GetUsersOrders(login)
	var sum float64 = 0
	for _, order := range orders {
		sum += order.Accrual
	}
	alreadyWithdrawn := db.GetWithdrawn(login)
	logger.Log(fmt.Sprintf("Sum in orders = %g, alreadyWothdrawn = %g, want to withdraw = %g", sum, alreadyWithdrawn, withdrawReq.Sum))
	if sum - alreadyWithdrawn < withdrawReq.Sum {
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}

	db.Withdraw(login, withdrawReq.Order, withdrawReq.Sum)
	w.WriteHeader(http.StatusOK)
}

func getWithdrawals(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value(userLoginKey{}).(string)
	withdrawals := db.GetAllWithdrawals(login)
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	body, _ := json.Marshal(withdrawals)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)

}

func mockedGetOrderStatus(w http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")
	order := models.Order{Number: number, Status: "PROCESSED", Accrual: float64(455.34)}
	body, _ := json.Marshal(order)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func readBody(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	defer r.Body.Close()
	err := dec.Decode(dst)
	return err
}
