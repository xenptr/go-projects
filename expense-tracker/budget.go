package main

import (
	"fmt"
	"time"
)

func getBudgetForMonth(m time.Month) (Budget, bool) {
	year := time.Now().Year()
	for i := range store.db.Budgets {
		if store.db.Budgets[i].Year == year && store.db.Budgets[i].Month == m {
			return store.db.Budgets[i], true
		}
	}
	return Budget{}, false
}

func checkBudgetWarning(m time.Month) {
	spent := totalByMonth(m)
	budget, ok := getBudgetForMonth(m)
	if ok && spent > budget.Amount {
		fmt.Printf(
			"Warning: You are $%.2f over your %s budget (Budget: $%.2f, Spent: $%.2f).\n",
			spent-budget.Amount,
			m,
			budget.Amount,
			spent,
		)
	}
}
