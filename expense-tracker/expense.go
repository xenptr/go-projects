package main

import (
	"fmt"
	"strings"
	"time"
)

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

func nextID() int {
	if len(db.Expenses) < 1 {
		return 1
	}

	var maxID int

	for i := range db.Expenses {
		if db.Expenses[i].ID > maxID {
			maxID = db.Expenses[i].ID
		}
	}

	return maxID + 1
}

func find(id int) (int, error) {
	for i := range db.Expenses {
		if db.Expenses[i].ID == id {
			return i, nil
		}
	}

	return -1, fmt.Errorf("expense %d not found", id)
}

func total() float64 {
	var sum float64
	for i := range db.Expenses {
		sum += db.Expenses[i].Amount
	}
	return sum
}

func totalByMonth(m time.Month) float64 {
	var sum float64
	for i := range db.Expenses {
		if db.Expenses[i].Date.Year() == time.Now().Year() && db.Expenses[i].Date.Month() == m {
			sum += db.Expenses[i].Amount
		}
	}
	return sum
}

func filterByCategory(c string) []Expense {
	var filtered []Expense

	for i := range db.Expenses {
		if strings.EqualFold(db.Expenses[i].Category, c) {
			filtered = append(filtered, db.Expenses[i])
		}
	}

	return filtered
}

func filterByMonth(m time.Month) []Expense {
	var filtered []Expense

	for i := range db.Expenses {
		if db.Expenses[i].Date.Month() == m {
			filtered = append(filtered, db.Expenses[i])
		}
	}

	return filtered
}

func filterByCategoryAndMonth(c string, m time.Month) []Expense {
	var filtered []Expense

	for i := range db.Expenses {
		if strings.EqualFold(db.Expenses[i].Category, c) && db.Expenses[i].Date.Month() == m {
			filtered = append(filtered, db.Expenses[i])
		}
	}

	return filtered
}

func getBudgetForMonth(m time.Month) (Budget, bool) {
	year := time.Now().Year()

	for i := range db.Budgets {
		if db.Budgets[i].Year == year && db.Budgets[i].Month == m {
			return db.Budgets[i], true
		}
	}

	return Budget{}, false
}

func setBudgetForMonth(m time.Month, amount float64) {
	year := time.Now().Year()

	for i := range db.Budgets {
		if db.Budgets[i].Year == year && db.Budgets[i].Month == m {
			db.Budgets[i].Amount = amount
			return
		}
	}

	db.Budgets = append(db.Budgets, Budget{
		Year:   year,
		Month:  m,
		Amount: amount,
	})
}
