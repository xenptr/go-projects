package main

import "time"

type Expense struct {
	ID          int       `json:"id"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
}

type Budget struct {
	Year   int        `json:"year"`
	Month  time.Month `json:"month"`
	Amount float64    `json:"amount"`
}

type Database struct {
	Expenses []Expense `json:"expenses"`
	Budgets  []Budget  `json:"budgets"`
}
