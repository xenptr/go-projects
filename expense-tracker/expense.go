package main

import (
	"fmt"
	"strings"
	"time"
)

type Expense struct {
	ID          int `json:"id"`
	Date        time.Time
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
}

var expenses []Expense

func nextID() int {
	if len(expenses) < 1 {
		return 0
	}

	var maxID int

	for i := range expenses {
		if expenses[i].ID > maxID {
			maxID = expenses[i].ID
		}
	}

	return maxID + 1
}

func find(id int) (int, error) {
	for i := range expenses {
		if expenses[i].ID == id {
			return i, nil
		}
	}

	return -1, fmt.Errorf("expense %d not found", id)
}

func total() float64 {
	var sum float64
	for i := range expenses {
		sum += expenses[i].Amount
	}
	return sum
}

func totalByMonth(m time.Month) float64 {
	var sum float64
	for i := range expenses {
		if expenses[i].Date.Year() == time.Now().Year() && expenses[i].Date.Month() == m {
			sum += expenses[i].Amount
		}
	}
	return sum
}

func filterByCategory(c string) []Expense {
	var filtered []Expense

	for i := range expenses {
		if strings.EqualFold(expenses[i].Category, c) {
			filtered = append(filtered, expenses[i])
		}
	}

	return filtered
}

func filterByMonth(m time.Month) []Expense {
	var filtered []Expense

	for i := range expenses {
		if expenses[i].Date.Month() == m {
			filtered = append(filtered, expenses[i])
		}
	}

	return filtered
}

func filterByCatMonth(c string, m time.Month) []Expense {
	var filtered []Expense

	for i := range expenses {
		if strings.EqualFold(expenses[i].Category, c) && expenses[i].Date.Month() == m {
			filtered = append(filtered, expenses[i])
		}
	}

	return filtered
}
