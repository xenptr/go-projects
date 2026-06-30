package main

import (
	"fmt"
	"os"
	"time"
)

func validMonth(m int) bool {
	return m >= 1 && m <= 12
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

func runAdd(description, category string, amount float64) {
	if description == "" {
		die("--description is required")
	}
	if amount <= 0 {
		die("--amount must be greater than zero")
	}
	if category == "" {
		die("--category is required")
	}

	exp := Expense{
		ID:          nextID(),
		Date:        time.Now(),
		Description: description,
		Amount:      amount,
		Category:    category,
	}

	if err := store.AddExpense(exp); err != nil {
		die("%v", err)
	}

	checkBudgetWarning(exp.Date.Month())
	fmt.Printf("Expense added successfully (ID: %d)\n", exp.ID)
}

func runUpdate(id int, description, category string, amount float64) {
	if id == 0 {
		die("--id is required")
	}

	index, err := find(id)
	if err != nil {
		die("%v", err)
	}

	if err := store.UpdateExpense(index, description, category, amount); err != nil {
		die("%v", err)
	}

	checkBudgetWarning(store.db.Expenses[index].Date.Month())
	fmt.Println("Expense updated successfully")
}

func runDelete(id int) {
	if id == 0 {
		die("--id is required")
	}

	index, err := find(id)
	if err != nil {
		die("%v", err)
	}

	if err := store.DeleteExpense(index); err != nil {
		die("%v", err)
	}

	fmt.Println("Expense deleted successfully")
}

func runList(category string) {
	expenses := store.db.Expenses

	if category != "" {
		expenses = filterByCategory(category)
		if len(expenses) == 0 {
			die("no expenses found for category: %s", category)
		}
	}

	printExpenseTable(expenses)
}

func printExpenseTable(expenses []Expense) {
	if len(expenses) == 0 {
		fmt.Println("No expenses found.")
		return
	}

	// Dynamically size the Description column to the longest entry.
	colDesc := len("Description")
	for _, exp := range expenses {
		if len(exp.Description) > colDesc {
			colDesc = len(exp.Description)
		}
	}

	// Header
	fmt.Printf("%-4s  %-12s  %-*s  %-10s  %-12s\n",
		"ID", "Date", colDesc, "Description", "Amount", "Category")
	// Separator
	fmt.Printf("%-4s  %-12s  %-*s  %-10s  %-12s\n",
		"----", "------------", colDesc, "------------", "----------", "------------")
	// Rows
	for _, exp := range expenses {
		fmt.Printf("%-4d  %-12s  %-*s  $%-9.2f  %-12s\n",
			exp.ID,
			exp.Date.Format(time.DateOnly),
			colDesc,
			exp.Description,
			exp.Amount,
			exp.Category,
		)
	}
}

func runSummary(month int) {
	if month != 0 {
		if !validMonth(month) {
			die("invalid month %d: must be between 1 and 12", month)
		}
		m := time.Month(month)
		fmt.Printf("Total expenses for %s: $%.2f\n", m, totalByMonth(m))
		return
	}

	fmt.Printf("Total expenses: $%.2f\n", total())
}

func runExport(category string, month int) {
	if month != 0 && !validMonth(month) {
		die("invalid month %d: must be between 1 and 12", month)
	}

	expenses := store.db.Expenses

	m := time.Month(month)
	switch {
	case category != "" && month != 0:
		expenses = filterByCategoryAndMonth(category, m)
	case category != "":
		expenses = filterByCategory(category)
	case month != 0:
		expenses = filterByMonth(m)
	}

	if err := store.ExportCSV(expenses); err != nil {
		die("%v", err)
	}

	fmt.Printf("Exported %d expense(s) to %s\n", len(expenses), exportFile)
}

func runBudget(month int, amount float64) {
	if month == 0 {
		die("--month is required")
	}
	if !validMonth(month) {
		die("invalid month %d: must be between 1 and 12", month)
	}

	m := time.Month(month)

	// Set a new budget when --amount is provided.
	if amount > 0 {
		spent := totalByMonth(m)

		if err := store.SetBudget(m, amount); err != nil {
			die("%v", err)
		}

		fmt.Printf("Budget set for %s: $%.2f\n", m, amount)
		if spent > amount {
			fmt.Printf(
				"Warning: Current spending ($%.2f) already exceeds this budget ($%.2f) by $%.2f.\n",
				spent, amount, spent-amount,
			)
		}
		return
	}

	// Show current spending status for the month.
	spent := totalByMonth(m)
	budget, ok := getBudgetForMonth(m)
	if !ok {
		fmt.Printf("No budget set for %s.\n", m)
		return
	}

	remaining := budget.Amount - spent
	fmt.Printf("Budget for %s: $%.2f\n", m, budget.Amount)
	fmt.Printf("Spent:         $%.2f\n", spent)
	if remaining >= 0 {
		fmt.Printf("Remaining:     $%.2f\n", remaining)
	} else {
		fmt.Printf("Over budget:   $%.2f\n", -remaining)
	}
}
