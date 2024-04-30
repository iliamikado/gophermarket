package main

import (
	"net/http"

	"github.com/iliamikado/gophermarket/internal/config"
	"github.com/iliamikado/gophermarket/internal/db"
	"github.com/iliamikado/gophermarket/internal/logger"
	"github.com/iliamikado/gophermarket/internal/router"
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
	logger.Log("Out api - " + config.AccrualSystemAddress)
	logger.Log("Start server on " + config.RunAddress)
	return http.ListenAndServe(config.RunAddress, r)
}
