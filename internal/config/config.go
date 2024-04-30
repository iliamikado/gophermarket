package config

import (
	"flag"
	"os"
)

var (
	RunAddress string
	DatabaseURI string
	AccrualSystemAddress string
	SecretKey string
)

func ParseConfig() {
	flag.StringVar(&RunAddress, "a", "localhost:8080", "Set launch address for server")
	flag.StringVar(&DatabaseURI, "d", "host=localhost user=demouser password=password dbname=gophermarket_db", "Set DB adress")
	flag.StringVar(&AccrualSystemAddress, "r", "http://localhost:8081", "Set accrual system address")
	flag.StringVar(&SecretKey, "s", "secret key", "Set secret key for coding")
	flag.Parse()

	if runAddress, exists := os.LookupEnv("RUN_ADDRESS"); exists {
		RunAddress = runAddress
	}
	if databaseURI, exists := os.LookupEnv("DATABASE_URI"); exists {
		DatabaseURI = databaseURI
	}
	if accrualSystemAddress, exists := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); exists {
		AccrualSystemAddress = accrualSystemAddress
	}
	if secretKey, exists := os.LookupEnv("SECRET_KEY"); exists {
		SecretKey = secretKey
	}
	AccrualSystemAddress += "/api/orders/"
	
}