package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	store = &Store{}
	if err := store.Load(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	addDescription := addCmd.String("description", "", "Expense description")
	addAmount := addCmd.Float64("amount", 0, "Expense amount")
	addCategory := addCmd.String("category", "", "Category name")

	updateCmd := flag.NewFlagSet("update", flag.ExitOnError)
	updateID := updateCmd.Int("id", 0, "Expense ID to update")
	updateDescription := updateCmd.String("description", "", "New description")
	updateAmount := updateCmd.Float64("amount", 0, "New amount")
	updateCategory := updateCmd.String("category", "", "New category name")

	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteID := deleteCmd.Int("id", 0, "Expense ID to delete")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listCategory := listCmd.String("category", "", "Filter by category")

	summaryCmd := flag.NewFlagSet("summary", flag.ExitOnError)
	summaryMonth := summaryCmd.Int("month", 0, "Month number (1-12)")

	exportCmd := flag.NewFlagSet("export", flag.ExitOnError)
	exportCategory := exportCmd.String("category", "", "Filter by category")
	exportMonth := exportCmd.Int("month", 0, "Filter by month number (1-12)")

	budgetCmd := flag.NewFlagSet("budget", flag.ExitOnError)
	budgetMonth := budgetCmd.Int("month", 0, "Month number (1-12)")
	budgetAmount := budgetCmd.Float64("amount", 0, "Budget amount (omit to view status)")

	usage := func() {
		fmt.Fprintln(os.Stderr, "Usage: expense-tracker <command> [options]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Commands:")
		fmt.Fprintln(os.Stderr, "  add      Add a new expense")
		fmt.Fprintln(os.Stderr, "  update   Update an existing expense")
		fmt.Fprintln(os.Stderr, "  delete   Delete an expense")
		fmt.Fprintln(os.Stderr, "  list     List expenses")
		fmt.Fprintln(os.Stderr, "  summary  Show expense totals")
		fmt.Fprintln(os.Stderr, "  export   Export expenses to CSV")
		fmt.Fprintln(os.Stderr, "  budget   Set or view a monthly budget")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Run 'expense-tracker <command> -help' for command-specific options.")
	}

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "add":
		addCmd.Parse(os.Args[2:])
		runAdd(*addDescription, *addCategory, *addAmount)

	case "update":
		updateCmd.Parse(os.Args[2:])
		runUpdate(*updateID, *updateDescription, *updateCategory, *updateAmount)

	case "delete":
		deleteCmd.Parse(os.Args[2:])
		runDelete(*deleteID)

	case "list":
		listCmd.Parse(os.Args[2:])
		runList(*listCategory)

	case "summary":
		summaryCmd.Parse(os.Args[2:])
		runSummary(*summaryMonth)

	case "export":
		exportCmd.Parse(os.Args[2:])
		runExport(*exportCategory, *exportMonth)

	case "budget":
		budgetCmd.Parse(os.Args[2:])
		runBudget(*budgetMonth, *budgetAmount)

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}
