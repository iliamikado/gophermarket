package main

import (
	"fmt"
	"net/http"

	"github.com/iliamikado/gophermarket/internal/router"
	"github.com/iliamikado/gophermarket/internal/db"
	"github.com/iliamikado/gophermarket/internal/config"
)

func main() {
	config.ParseConfig()

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	r := router.AppRouter()
	db.Initialize(config.DatabaseURI)
	fmt.Println("start server")
	return http.ListenAndServe(config.RunAddress, r)
}
