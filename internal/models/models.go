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

type Balance struct {
	Current		float64 `json:"current"`
	Withdrawn	float64	`json:"withdrawn"`
}

type WithdrawRequest struct {
	Order		string 	`json:"order"`
	Sum			float64 `json:"sum"`
}

type WithdrawLog struct {
	Order		string	`json:"order"`
	Sum			float64	`json:"sum"`
	Date		string	`json:"processed_at"`
}