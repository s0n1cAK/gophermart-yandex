package models

import "time"

type User struct {
	ID       uint64 `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	UserID     uint64     `json:"user_id"`
	Number     string     `json:"number"`
	Status     string     `json:"status"`
	Accrual    float64    `json:"accrual,omitempty"`
	UploadedAt *time.Time `json:"uploaded_at,omitempty"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Withdrawal struct {
	UserID      string     `json:"user_id"`
	Order       string     `json:"order"`
	Sum         float64    `json:"sum"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}
