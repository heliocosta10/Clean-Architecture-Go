package domain

import "time"

type Order struct {
	ID          int64     `json:"id"`
	Customer    string    `json:"customer"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
}
