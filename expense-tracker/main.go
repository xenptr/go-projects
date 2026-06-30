package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	if err := load(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	addFlags := flag.NewFlagSet("add", flag.ExitOnError)
	addDescription := addFlags.String("description", "", "Expense description")
	addAmount := addFlags.Float64("amount", 0, "Expense amount")
	addCategory := addFlags.String("category", "", "Category name")

	updateFlags := flag.NewFlagSet("update", flag.ExitOnError)
	updateID := updateFlags.Int("id", 0, "Expense ID")
	updateDescription := updateFlags.String("description", "", "New description")
	updateAmount := updateFlags.Float64("amount", 0, "New amount")
	updateCategory := updateFlags.String("category", "", "Category name")

	deleteFlags := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteID := deleteFlags.Int("id", 0, "Expense ID")

	listFlags := flag.NewFlagSet("list", flag.ExitOnError)
	listCategory := listFlags.String("category", "", "Category name")

	summaryFlags := flag.NewFlagSet("summary", flag.ExitOnError)
	summaryMonth := summaryFlags.Int("month", 0, "Month number (1-12)")

	exportFlags := flag.NewFlagSet("export", flag.ExitOnError)
	exportCategory := exportFlags.String("category", "", "Category name")
	exportMonth := exportFlags.Int("month", 0, "Month number (1-12)")

	budgetFlags := flag.NewFlagSet("budget", flag.ExitOnError)
	budgetMonth := budgetFlags.Int("month", 0, "Month number (1-12)")
	budgetAmount := budgetFlags.Float64("amount", 0, "Budget amount")

	flag.Usage = func() {
		fmt.Println("Usage:")
		fmt.Println("  expense-tracker <command> [options]")

		fmt.Fprintln(os.Stderr, "Commands:")
		fmt.Fprintln(os.Stderr, "  add       Add a new expense")
		fmt.Fprintln(os.Stderr, "  update    Update an existing expense")
		fmt.Fprintln(os.Stderr, "  delete    Delete an expense")
		fmt.Fprintln(os.Stderr, "  list      List expenses")
		fmt.Fprintln(os.Stderr, "  summary   Show expense totals")
		fmt.Fprintln(os.Stderr, "  export    Export expenses to CSV")
		fmt.Fprintln(os.Stderr, "  budget    Set or view monthly budgets")

		for _, fs := range []*flag.FlagSet{
			addFlags,
			updateFlags,
			deleteFlags,
			listFlags,
			summaryFlags,
			exportFlags,
			budgetFlags,
		} {
			fmt.Printf("\n%s:\n", fs.Name())
			fs.PrintDefaults()
		}
	}

	flag.Parse()

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	expenses := db.Expenses

	switch os.Args[1] {
	case "add":
		addFlags.Parse(os.Args[2:])

		if *addDescription == "" {
			fmt.Println("description is required")
			return
		}

		if *addAmount <= 0 {
			fmt.Println("amount must be greater than zero")
			return
		}

		if *addCategory == "" {
			fmt.Println("category is required")
			return
		}

		data := Expense{
			ID:          nextID(),
			Date:        time.Now(),
			Description: *addDescription,
			Amount:      *addAmount,
			Category:    *addCategory,
		}
		db.Expenses = append(db.Expenses, data)
		if err := save(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		spent := totalByMonth(time.Now().Month())
		if budget, ok := getBudgetForMonth(time.Now().Month()); ok && spent > budget.Amount {
			fmt.Printf(
				"Warning: You are %.2f over your %s budget (Budget: %.2f, Spent: %.2f).\n",
				spent-budget.Amount,
				time.Now().Month(),
				budget.Amount,
				spent,
			)
		}

		fmt.Printf("Expense added successfully (ID: %d)\n", data.ID)
		return

	case "update":
		updateFlags.Parse(os.Args[2:])

		if *updateID == 0 {
			fmt.Println("id is required")
			return
		}

		if *updateAmount <= 0 {
			fmt.Println("amount must be greater than zero")
			return
		}

		index, err := find(*updateID)
		if err != nil {
			fmt.Printf("update expense: %v\n", err)
			return
		}

		if *updateDescription != "" {
			db.Expenses[index].Description = *updateDescription
		}
		if *updateAmount > 0 {
			db.Expenses[index].Amount = *updateAmount
		}
		if *updateCategory != "" {
			db.Expenses[index].Category = *updateCategory
		}
		db.Expenses[index].Date = time.Now()

		if err := save(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		spent := totalByMonth(db.Expenses[index].Date.Month())
		if budget, ok := getBudgetForMonth(db.Expenses[index].Date.Month()); ok && spent > budget.Amount {
			fmt.Printf(
				"Warning: You are %.2f over your %s budget (Budget: %.2f, Spent: %.2f).\n",
				spent-budget.Amount,
				db.Expenses[index].Date.Month(),
				budget.Amount,
				spent,
			)
		}

		fmt.Println("Expense updated successfully")
		return

	case "delete":
		deleteFlags.Parse(os.Args[2:])

		index, err := find(*deleteID)
		if err != nil {
			fmt.Printf("delete expense: %v\n", err)
			return
		}
		db.Expenses = append(db.Expenses[:index], db.Expenses[index+1:]...)
		if err := save(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Expense deleted successfully")
		return

	case "list":
		listFlags.Parse(os.Args[2:])

		if *listCategory != "" {
			expenses = filterByCategory(*listCategory)
			if len(expenses) == 0 {
				fmt.Printf("%s category not found", *listCategory)
				return
			}
		}

		maxDesc := len("Description")
		for _, exp := range expenses {
			if len(exp.Description) > maxDesc {
				maxDesc = len(exp.Description)
			}
		}

		fmt.Printf("%-4s %-12s %-*s %-10s %-12s\n", "ID", "Date", maxDesc+1, "Description", "Amount", "Category")
		for _, exp := range expenses {
			amount := fmt.Sprintf("$%.2f", exp.Amount)
			fmt.Printf("%-4d %-12s %-*s %-10s %-12s\n",
				exp.ID,
				exp.Date.Format(time.DateOnly),
				maxDesc+1,
				exp.Description,
				amount,
				exp.Category,
			)
		}
		return

	case "summary":
		summaryFlags.Parse(os.Args[2:])

		if *summaryMonth != 0 {
			fmt.Printf("Total expenses for %s: $%.2f\n",
				time.Month(*summaryMonth).String(),
				totalByMonth(time.Month(*summaryMonth)),
			)
			return
		}

		fmt.Printf("Total expenses: $%.2f\n", total())
		return

	case "export":
		exportFlags.Parse(os.Args[2:])

		if *exportCategory != "" && *exportMonth != 0 {
			expenses = filterByCategoryAndMonth(*exportCategory, time.Month(*exportMonth))
		} else if *exportCategory != "" {
			expenses = filterByCategory(*exportCategory)
		} else if *exportMonth != 0 {
			expenses = filterByMonth(time.Month(*exportMonth))
		}

		if err := export(expenses); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return

	case "budget":
		budgetFlags.Parse(os.Args[2:])

		if *budgetMonth != 0 && *budgetAmount > 0 {
			spent := totalByMonth(time.Month(*budgetMonth))

			setBudgetForMonth(time.Month(*budgetMonth), *budgetAmount)

			if spent > *budgetAmount {
				fmt.Printf(
					"Warning: Current spending (%.2f) already exceeds the budget you just set (%.2f) by %.2f.\n",
					spent,
					*budgetAmount,
					spent-*budgetAmount,
				)
			}

			return
		}

		if *budgetMonth != 0 {
			spent := totalByMonth(time.Month(*budgetMonth))

			if budget, ok := getBudgetForMonth(time.Month(*budgetMonth)); ok {
				fmt.Printf("Spent: %.2f / %.2f\n", spent, budget.Amount)
				fmt.Printf("Remaining: %.2f\n", budget.Amount-spent)
			} else {
				fmt.Println("No budget has been set for this month")
			}
		}

		return

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		flag.Usage()
		os.Exit(1)
	}
}
