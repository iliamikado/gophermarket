package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func Initialize(host string) {
	db, _ := sql.Open("pgx", host)
	DB = db
	CreateTables()
}

func CreateTables() {
	_, err := DB.Exec(`CREATE TABLE IF NOT EXISTS users (
		login TEXT PRIMARY KEY NOT NULL,
		password TEXT NOT NULL
	)`)
	fmt.Println(err)
}

func IsLoginExist(login string) bool {
	rows, _ := DB.Query(`SELECT * FROM users WHERE login = $1`, login)
	defer rows.Close()
	return rows.Next()
}

func AddNewUser(login, password string) {
	passwordHash := password
	DB.Exec(`INSERT INTO users VALUES ($1, $2)`, login, passwordHash)
}

func IsValidUser(login, password string) bool {
	row := DB.QueryRow(`SELECT password FROM users WHERE login = $1`, login)
	if (row == nil) {
		return false
	}
	var passwordHash string
	row.Scan(&passwordHash)
	return password == passwordHash
}