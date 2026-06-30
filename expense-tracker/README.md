# Expense Tracker CLI

A small CLI tool to manage your personal finances from the terminal. Add expenses, track spending by category, set monthly budgets, and export your data to CSV.

## Project URL

https://roadmap.sh/projects/expense-tracker

## Features

- Add, update, and delete expenses
- List all expenses with a formatted table
- Filter expenses by category
- View total spending or a monthly summary
- Set monthly budgets with over-budget warnings
- Export expenses to CSV with optional filters

## Installation

Clone the repository:

```bash
git clone <repository-url>
cd expense-tracker
```

Run without building:

```bash
go run . <command> [options]
```

Or build an executable:

```bash
go build -o expense-tracker
```

Run the executable:

```bash
./expense-tracker <command> [options]
```

## Usage

```
expense-tracker <command> [options]
```

| Command | Description |
|---------|-------------|
| `add` | Add a new expense |
| `update` | Update an existing expense |
| `delete` | Delete an expense |
| `list` | List expenses |
| `summary` | Show expense totals |
| `export` | Export expenses to CSV |
| `budget` | Set or view a monthly budget |

Run `expense-tracker <command> -help` for command-specific options.

### add

| Flag | Description |
|------|-------------|
| `--description <text>` | Expense description (required) |
| `--amount <number>` | Expense amount (required, must be > 0) |
| `--category <name>` | Category name (required) |

### update

| Flag | Description |
|------|-------------|
| `--id <number>` | ID of the expense to update (required) |
| `--description <text>` | New description |
| `--amount <number>` | New amount |
| `--category <name>` | New category |

### delete

| Flag | Description |
|------|-------------|
| `--id <number>` | ID of the expense to delete (required) |

### list

| Flag | Description |
|------|-------------|
| `--category <name>` | Filter by category (optional) |

### summary

| Flag | Description |
|------|-------------|
| `--month <1-12>` | Show total for a specific month (current year) |

### export

| Flag | Description |
|------|-------------|
| `--category <name>` | Filter by category |
| `--month <1-12>` | Filter by month (current year) |

### budget

| Flag | Description |
|------|-------------|
| `--month <1-12>` | Month to set or view (required) |
| `--amount <number>` | Budget limit (omit to view current status) |

Examples:

```bash
./expense-tracker add --description "Lunch" --amount 20 --category food
./expense-tracker add --description "Bus pass" --amount 45 --category travel
./expense-tracker list
./expense-tracker list --category food
./expense-tracker update --id 1 --amount 25
./expense-tracker delete --id 2
./expense-tracker summary
./expense-tracker summary --month 6
./expense-tracker budget --month 6 --amount 500
./expense-tracker budget --month 6
./expense-tracker export
./expense-tracker export --category food
./expense-tracker export --month 6
```

## Project Structure

```text
.
├── main.go         # Entry point, flag setup, and command routing
├── commands.go     # Command handlers (runAdd, runList, etc.)
├── expense.go      # Expense queries and filters
├── budget.go       # Budget logic and over-budget warnings
├── storage.go      # Store type, persistence (load/save), and CSV export
├── types.go        # Expense, Budget, and Database structs
├── expenses.json   # Data file (created automatically on first use)
├── go.mod
└── README.md
```

## Data Storage

Expenses and budgets are stored locally in `expenses.json` in the working directory. The file is created automatically the first time you add an expense. Budgets are tracked per month and year.
