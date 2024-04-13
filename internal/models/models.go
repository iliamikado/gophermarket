package models

type User struct {
	Login 		string	`json:"login"`
	Password 	string	`json:"password"`
}

type Order struct {
	Number		string	`json:"number"`
	Date 		string	`json:"uploaded_at"`
	Status 		string	`json:"status"`
	Accrual 	float64	`json:"accrual,omitempty"`
}