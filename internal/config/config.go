package config

import (
	"flag"
	"os"
)

var (
	RunAddress string
	DatabaseURI string
	AccrualSystemAddress string
)

func ParseConfig() {
	flag.StringVar(&RunAddress, "a", "localhost:8080", "Set launch address for server")
	flag.StringVar(&DatabaseURI, "d", "host=localhost user=demouser password=password dbname=gophermarket_db", "Set DB adress")
	flag.StringVar(&AccrualSystemAddress, "r", "http://localhost:8080/mock/", "Set accural system address")
	flag.Parse()

	if runAddress := os.Getenv("RUN_ADDRESS"); runAddress != "" {
		RunAddress = runAddress
	}
	if databaseURI := os.Getenv("DATABASE_URI"); databaseURI != "" {
		DatabaseURI = databaseURI
	}
	if accrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); accrualSystemAddress != "" {
		AccrualSystemAddress = accrualSystemAddress
	}
	
}