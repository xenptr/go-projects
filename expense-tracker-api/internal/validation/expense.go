package validation

import (
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/dto"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/models"
)

// ValidateCreateExpense validates and normalises a create-expense request.
func ValidateCreateExpense(req *dto.CreateExpenseRequest) error {
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return errors.New("title is required")
	}

	if req.Amount == "" {
		return errors.New("amount is required")
	}
	amt, err := decimal.NewFromString(req.Amount)
	if err != nil || !amt.IsPositive() {
		return errors.New("amount must be a positive number")
	}

	if req.Category == "" {
		req.Category = models.CategoryOthers
	}
	if !req.Category.Valid() {
		return errors.New("category is invalid")
	}

	if req.Date == "" {
		req.Date = time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		return errors.New("date must be in YYYY-MM-DD format")
	}

	return nil
}

// ValidateUpdateExpense validates an update-expense request (all fields optional).
func ValidateUpdateExpense(req *dto.UpdateExpenseRequest) error {
	if req.Title != nil {
		*req.Title = strings.TrimSpace(*req.Title)
		if *req.Title == "" {
			return errors.New("title cannot be empty")
		}
	}

	if req.Amount != nil {
		amt, err := decimal.NewFromString(*req.Amount)
		if err != nil || !amt.IsPositive() {
			return errors.New("amount must be a positive number")
		}
	}

	if req.Category != nil && !req.Category.Valid() {
		return errors.New("category is invalid")
	}

	if req.Date != nil {
		if _, err := time.Parse("2006-01-02", *req.Date); err != nil {
			return errors.New("date must be in YYYY-MM-DD format")
		}
	}

	return nil
}
