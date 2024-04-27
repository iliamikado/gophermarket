package db

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/iliamikado/gophermarket/internal/config"
	"github.com/iliamikado/gophermarket/internal/logger"
	"github.com/iliamikado/gophermarket/internal/models"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB
var UserAlreadyExistsError error
var UserAlreadyHasOrderError error
var AnotherUserAlreadyHasOrderError error
var NotEnoughPointsError error

func Initialize(host string) {
	db, _ := sql.Open("pgx", host)
	DB = db
	CreateTables()

	UserAlreadyExistsError = errors.New("user with this login already exists")
	UserAlreadyHasOrderError = errors.New("user already has this order")
	AnotherUserAlreadyHasOrderError = errors.New("another user already has this order")
	NotEnoughPointsError = errors.New("not enough points")
}

func CreateTables() {
	DB.Exec(`CREATE TABLE IF NOT EXISTS users (
		login TEXT PRIMARY KEY NOT NULL,
		password BYTEA NOT NULL,
		withdrawn DECIMAL NOT NULL DEFAULT 0
	)`)
	DB.Exec(`CREATE TABLE IF NOT EXISTS orders (
		id TEXT PRIMARY KEY NOT NULL,
		status TEXT NOT NULL DEFAULT 'NEW',
		accrual DECIMAL,
		user_login TEXT REFERENCES users (login),
		date TIMESTAMP NOT NULL DEFAULT NOW()
	)`)
	DB.Exec(`CREATE TABLE IF NOT EXISTS withdrawals (
		order_number TEXT,
		sum DECIMAL,
		user_login TEXT REFERENCES users (login),
		processed_at TIMESTAMP NOT NULL DEFAULT NOW()
	)`)
}

func AddNewUser(login, password string) error {
	tx, _ := DB.Begin()

	row := tx.QueryRow(`SELECT * FROM users WHERE login = $1`, login)
	err := row.Scan()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		tx.Rollback()
		return err
	} else if err == nil {
		tx.Rollback()
		return UserAlreadyExistsError
	}

	h := sha256.New()
	h.Write([]byte(password + config.SecretKey))
	passwordHash := h.Sum(nil)
	logger.Log(passwordHash)
	_, err = tx.Exec(`INSERT INTO users VALUES ($1, $2)`, login, passwordHash)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func IsValidUser(login, password string) bool {
	row := DB.QueryRow(`SELECT password FROM users WHERE login = $1`, login)
	if row == nil {
		return false
	}
	var passwordHash []byte
	row.Scan(&passwordHash)
	h := sha256.New()
	h.Write([]byte(password + config.SecretKey))
	return bytes.Equal(h.Sum(nil), passwordHash)
}

func AddNewOrder(order models.Order, login string) error {
	tx, _ := DB.Begin()
	row := tx.QueryRow(`SELECT user_login FROM orders WHERE id = $1`, order.Number)
	var orderLogin string
	err := row.Scan(&orderLogin)
	if err == nil {
		tx.Rollback()
		if orderLogin == login {
			return UserAlreadyHasOrderError
		} else {
			return AnotherUserAlreadyHasOrderError
		}
	}
	if !errors.Is(err, sql.ErrNoRows) {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec(`INSERT INTO orders (id, status, accrual, user_login) VALUES ($1, $2, $3, $4)`, order.Number, order.Status, order.Accrual, login)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func UpdateOrder(order models.Order) {
	DB.Exec(`UPDATE orders SET (status, accrual) = ($1, $2) WHERE id = $3`, order.Status, order.Accrual, order.Number)
}

func GetUsersOrders(login string) []models.Order {
	rows, err := DB.Query(`SELECT id, status, accrual, date FROM orders WHERE user_login = $1 ORDER BY date`, login)
	if err != nil {
		logger.Log(err)
	}
	ans := make([]models.Order, 0)
	for rows.Next() {
		var number, status string
		var accrual float64
		var date time.Time
		rows.Scan(&number, &status, &accrual, &date)
		ans = append(ans, models.Order{Number: number, Status: status, Accrual: accrual, Date: date.Format(time.RFC3339)})
	}
	if err := rows.Err(); err != nil {
		logger.Log(err)
		panic(err)
	}
	return ans
}

func Withdraw(login, order string, amount float64) error {
	tx, _ := DB.Begin()
	row := tx.QueryRow(`SELECT SUM(accrual) FROM orders WHERE user_login = $1 group by user_login`, login)
	var sum float64
	row.Scan(&sum)
	row = tx.QueryRow(`SELECT withdrawn FROM users WHERE login = $1`, login)
	var withdrawn float64
	row.Scan(&withdrawn)

	if sum - withdrawn < amount {
		tx.Rollback()
		return NotEnoughPointsError
	}

	_, e1 := tx.Exec(`UPDATE users SET withdrawn = withdrawn + $1 WHERE login = $2`, amount, login)
	_, e2 := tx.Exec(`INSERT INTO withdrawals (order_number, sum, user_login) VALUES ($1, $2, $3)`, order, amount, login)
	if e1 != nil || e2 != nil {
		logger.Log(e1)
		logger.Log(e2)
		tx.Rollback()
	}
	return tx.Commit()
}

func GetBalance(login string) (float64, float64) {
	tx, _ := DB.Begin()
	row := tx.QueryRow(`SELECT SUM(accrual) FROM orders WHERE user_login = $1 group by user_login`, login)
	var sum float64
	row.Scan(&sum)
	row = tx.QueryRow(`SELECT withdrawn FROM users WHERE login = $1`, login)
	var withdrawn float64
	row.Scan(&withdrawn)
	tx.Commit()
	return sum, withdrawn
}

func GetAllWithdrawals(login string) []models.WithdrawLog {
	rows, err := DB.Query(`SELECT order_number, sum, processed_at FROM withdrawals WHERE user_login = $1 ORDER BY processed_at`, login)
	if err != nil {
		logger.Log(err)
	}
	ans := make([]models.WithdrawLog, 0)
	for rows.Next() {
		var order string
		var sum float64
		var date time.Time
		rows.Scan(&order, &sum, &date)
		ans = append(ans, models.WithdrawLog{Order: order, Sum: sum, Date: date.Format(time.RFC3339)})
	}
	if err := rows.Err(); err != nil {
		logger.Log(err)
		panic(err)
	}
	return ans
}
