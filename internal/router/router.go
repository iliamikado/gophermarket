package router

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/iliamikado/gophermarket/internal/db"
	"github.com/iliamikado/gophermarket/internal/logger"
	"github.com/iliamikado/gophermarket/internal/models"
)

func AppRouter() *chi.Mux{
	r := chi.NewRouter()
	r.Post("/api/user/register", register)
	r.Post("/api/user/login", login)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/api/user/orders", postOrder)
		r.Get("/api/user/orders", getOrders)
		r.Get("/api/user/balance", getBalance)
		r.Post("/api/user/balance/withdraw", pointsWithdraw)
		r.Get("/api/user/withdrawals", getWithdawals)
	})
	r.Get("/mock/{number}", mockedGetOrderStatus)
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
	if (db.IsLoginExist(user.Login)) {
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
	if (!db.IsValidUser(user.Login, user.Password)) {
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
	if (err != nil) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ordersUserLogin, exists := db.FindOrder(number)
	if (exists) {
		if (ordersUserLogin == login) {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusConflict)
		}
		return
	}
	db.AddNewOrder(models.Order{Number: number}, login)
	go func() {
		getOrderInfo(number)
	}()
	w.WriteHeader(http.StatusAccepted)
}

func getOrders(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value(userLoginKey{}).(string)
	orders := db.GetUsersOrders(login)
	logger.Log(orders)
	msg := ""
	for _, order := range orders {
		msg += "(" + order.Number + ", " + order.Status + ") "
	}
	logger.Log("Get orders from login " + login + ": " + msg)
	w.Header().Set("Content-Type", "application/json")
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	wg := sync.WaitGroup{}
	for i, order := range orders {
		if (order.Status == "INVALID" || order.Status == "PROCESSED") {
			continue
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			order := getOrderInfo(orders[i].Number)
			orders[i].Status = order.Status
			orders[i].Accural = order.Accural
		}(i)
	}
	wg.Wait()
	resp, _ := json.Marshal(orders)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func getBalance(w http.ResponseWriter, r *http.Request) {
	
}

func pointsWithdraw(w http.ResponseWriter, r *http.Request) {
	
}

func getWithdawals(w http.ResponseWriter, r *http.Request) {
	
}

func mockedGetOrderStatus(w http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")
	order := models.Order{Number: number, Status: "PROCESSED", Accural: 500}
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
