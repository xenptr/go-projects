package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Category string

const (
	CategoryGroceries   Category = "Groceries"
	CategoryLeisure     Category = "Leisure"
	CategoryElectronics Category = "Electronics"
	CategoryUtilities   Category = "Utilities"
	CategoryClothing    Category = "Clothing"
	CategoryHealth      Category = "Health"
	CategoryOthers      Category = "Others"
)

func (c Category) Valid() bool {
	switch c {
	case CategoryGroceries, CategoryLeisure, CategoryElectronics,
		CategoryUtilities, CategoryClothing, CategoryHealth, CategoryOthers:
		return true
	}
	return false
}

type Expense struct {
	ID          int64           `json:"id"`
	UserID      int64           `json:"user_id"`
	Title       string          `json:"title"`
	Amount      decimal.Decimal `json:"amount"`
	Category    Category        `json:"category"`
	Date        time.Time       `json:"date"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
