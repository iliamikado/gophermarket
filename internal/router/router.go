package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/iliamikado/gophermarket/internal/models"
	"github.com/iliamikado/gophermarket/internal/db"
)

func AppRouter() *chi.Mux{
	r := chi.NewRouter()
	r.Post("/api/user/register", register)
	r.Post("/api/user/login", login)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/api/user/orders", postOrder)
		r.Get("/api/user/orders", getOrder)
		r.Get("/api/user/balance", getBalance)
		r.Post("/api/user/balance/withdraw", pointsWithdraw)
		r.Get("/api/user/withdrawals", getWithdawals)
	})
	return r
}

func register(w http.ResponseWriter, r *http.Request) {
	var user models.User

	err := readBody(r, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(user)
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
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(user)
	if (!db.IsValidUser(user.Login, user.Password)) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "JWT", Value: buildJWTString(user.Login), Path: "/"})
	w.WriteHeader(http.StatusOK)
}

func postOrder(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value(userLoginKey{}).(string)
	fmt.Println(login)
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	
}

func getBalance(w http.ResponseWriter, r *http.Request) {
	
}

func pointsWithdraw(w http.ResponseWriter, r *http.Request) {
	
}

func getWithdawals(w http.ResponseWriter, r *http.Request) {
	
}

func readBody(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	defer r.Body.Close()
	err := dec.Decode(dst)
	return err
}