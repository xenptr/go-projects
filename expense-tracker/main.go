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
	addCategory := addFlags.String("category", "", "Expense category")

	updateFlags := flag.NewFlagSet("update", flag.ExitOnError)
	updateID := updateFlags.Int("id", 0, "Expense ID")
	updateDescription := updateFlags.String("description", "", "New description")
	updateAmount := updateFlags.Float64("amount", 0, "New amount")
	updateCategory := updateFlags.String("category", "", "Expense category")

	deleteFlags := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteID := deleteFlags.Int("id", 0, "Expense ID")

	listFlags := flag.NewFlagSet("list", flag.ExitOnError)
	listCategory := listFlags.String("category", "", "Expense category")

	summaryFlags := flag.NewFlagSet("summary", flag.ExitOnError)
	summaryMonth := summaryFlags.Int("month", 0, "Month")

	exportFlags := flag.NewFlagSet("export", flag.ExitOnError)
	exportCategory := exportFlags.String("category", "", "Expense category")
	exportMonth := exportFlags.Int("month", 0, "Month")

	flag.Usage = func() {
		fmt.Println("Usage:")
		fmt.Println("  expense-tracker <command> [options]")

		fmt.Fprintln(os.Stderr, "Commands:")
		fmt.Fprintln(os.Stderr, "  add       Add a new expense")
		fmt.Fprintln(os.Stderr, "  update    Update an expense")
		fmt.Fprintln(os.Stderr, "  delete    Delete an expense")
		fmt.Fprintln(os.Stderr, "  list      List expenses")
		fmt.Fprintln(os.Stderr, "  summary   Show total expenses")
		fmt.Fprintln(os.Stderr, "  export    Export expenses to CSV")

		for _, fs := range []*flag.FlagSet{
			addFlags,
			updateFlags,
			deleteFlags,
			listFlags,
		} {
			fmt.Printf("\n%s:\n", fs.Name())
			fs.PrintDefaults()
		}
	}

	flag.Parse()

	if len(os.Args) < 2 {
		os.Exit(1)
	}

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
		expenses = append(expenses, data)
		if err := save(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Expense added successfully (ID: %d)\n", data.ID)
		return

	case "update":
		updateFlags.Parse(os.Args[2:])

		if *updateAmount <= 0 {
			fmt.Println("amount must be greater than zero")
			return
		}

		index, err := find(*updateID)
		if err != nil {
			fmt.Printf("update expense: %v\n", err)
			return
		}

		expenses[index].Date = time.Now()
		expenses[index].Description = *updateDescription
		expenses[index].Amount = *updateAmount
		expenses[index].Category = *updateCategory

		if err := save(); err != nil {
			fmt.Println(err)
			os.Exit(1)
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
		expenses = append(expenses[:index], expenses[index+1:]...)
		if err := save(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
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
			fmt.Printf("Total expenses for %s: $%.2f",
				time.Month(*summaryMonth).String(),
				totalByMonth(time.Month(*summaryMonth)),
			)
			return
		}

		fmt.Printf("Total expenses: $%.2f", total())
		return

	case "export":
		exportFlags.Parse(os.Args[2:])

		if *exportCategory != "" {
			expenses = filterByCategory(*exportCategory)
		}

		if *exportMonth != 0 {
			expenses = filterByMonth(time.Month(*exportMonth))
		}

		if *exportCategory != "" && *exportMonth != 0 {
			expenses = filterByCatMonth(*exportCategory, time.Month(*exportMonth))
		}

		if err := export(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	default:
		return
	}
}
