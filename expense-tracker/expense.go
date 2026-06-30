package main

import (
	"fmt"
	"strings"
	"time"
)

func nextID() int {
	if len(store.db.Expenses) == 0 {
		return 1
	}
	var maxID int
	for _, exp := range store.db.Expenses {
		if exp.ID > maxID {
			maxID = exp.ID
		}
	}
	return maxID + 1
}

func find(id int) (int, error) {
	for i := range store.db.Expenses {
		if store.db.Expenses[i].ID == id {
			return i, nil
		}
	}
	return -1, fmt.Errorf("expense %d not found", id)
}

func total() float64 {
	var sum float64
	for _, exp := range store.db.Expenses {
		sum += exp.Amount
	}
	return sum
}

func totalByMonth(m time.Month) float64 {
	var sum float64
	for _, exp := range store.db.Expenses {
		if exp.Date.Year() == time.Now().Year() && exp.Date.Month() == m {
			sum += exp.Amount
		}
	}
	return sum
}

func filterByCategory(c string) []Expense {
	var filtered []Expense
	for _, exp := range store.db.Expenses {
		if strings.EqualFold(exp.Category, c) {
			filtered = append(filtered, exp)
		}
	}
	return filtered
}

func filterByMonth(m time.Month) []Expense {
	var filtered []Expense
	for _, exp := range store.db.Expenses {
		if exp.Date.Year() == time.Now().Year() && exp.Date.Month() == m {
			filtered = append(filtered, exp)
		}
	}
	return filtered
}

func filterByCategoryAndMonth(c string, m time.Month) []Expense {
	var filtered []Expense
	for _, exp := range store.db.Expenses {
		if strings.EqualFold(exp.Category, c) && exp.Date.Year() == time.Now().Year() && exp.Date.Month() == m {
			filtered = append(filtered, exp)
		}
	}
	return filtered
}
