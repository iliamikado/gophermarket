package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/iliamikado/gophermarket/internal/models"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func Initialize(host string) {
	db, _ := sql.Open("pgx", host)
	DB = db
	CreateTables()
}

func CreateTables() {
	DB.Exec(`CREATE TABLE IF NOT EXISTS users (
		login TEXT PRIMARY KEY NOT NULL,
		password TEXT NOT NULL
	)`)
	DB.Exec(`CREATE TABLE IF NOT EXISTS orders (
		id TEXT PRIMARY KEY NOT NULL,
		user_login TEXT REFERENCES users (login),
		date TIMESTAMP NOT NULL DEFAULT NOW()
	)`)
}

func IsLoginExist(login string) bool {
	row := DB.QueryRow(`SELECT * FROM users WHERE login = $1`, login)
	err := row.Scan()
	return !(err != nil && errors.Is(err, sql.ErrNoRows))
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

func AddNewOrder(orderNumber, login string) {
	DB.Exec(`INSERT INTO orders (id, user_login) VALUES ($1, $2)`, orderNumber, login)
}

func FindOrder(orderNumber string) (string, bool) {
	row := DB.QueryRow(`SELECT user_login FROM orders WHERE id = $1`, orderNumber)
	var login string
	err := row.Scan(&login)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return "", false
	}
	return login, true
}

func GetUsersOrders(login string) []models.Order {
	rows, _ := DB.Query(`SELECT id, date FROM orders WHERE user_login = $1 ORDER BY date`, login)
	ans := make([]models.Order, 0)
	for rows.Next() {
		var number string
		var date time.Time
		rows.Scan(&number, &date)
		ans = append(ans, models.Order{Number: number, Date: date.Format(time.RFC3339)})
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	return ans
}