package dto

import "github.com/xenptr/go-projects/expense-tracker-api/internal/models"

// CreateExpenseRequest is the body for POST /expenses.
type CreateExpenseRequest struct {
	Title       string          `json:"title"`
	Amount      string          `json:"amount"` // decimal string, e.g. "12.50"
	Category    models.Category `json:"category"`
	Date        string          `json:"date"`        // RFC3339 or YYYY-MM-DD
	Description string          `json:"description"` // optional
}

// UpdateExpenseRequest is the body for PUT /expenses/{id}.
// All fields are optional — only non-zero values are applied.
type UpdateExpenseRequest struct {
	Title       *string          `json:"title"`
	Amount      *string          `json:"amount"`
	Category    *models.Category `json:"category"`
	Date        *string          `json:"date"`
	Description *string          `json:"description"`
}

// ExpenseFilterRequest holds query params for GET /expenses.
type ExpenseFilterRequest struct {
	Filter    string `json:"filter"`     // "week" | "month" | "3months" | "custom"
	StartDate string `json:"start_date"` // YYYY-MM-DD, used when filter=custom
	EndDate   string `json:"end_date"`   // YYYY-MM-DD, used when filter=custom
}
