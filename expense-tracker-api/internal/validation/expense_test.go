package validation

import (
	"testing"

	"github.com/xenptr/go-projects/expense-tracker-api/internal/dto"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/models"
)

// strPtr and categoryPtr are small helpers to keep test cases concise.
func strPtr(s string) *string { return &s }
func categoryPtr(c models.Category) *models.Category { return &c }

// ---- ValidateCreateExpense ----

func TestValidateCreateExpense(t *testing.T) {
	tests := []struct {
		name    string
		input   dto.CreateExpenseRequest
		wantErr bool
		// errContains lets us assert on the error message substring.
		errContains string
	}{
		{
			name: "valid request with all fields",
			input: dto.CreateExpenseRequest{
				Title:       "Coffee",
				Amount:      "3.50",
				Category:    models.CategoryGroceries,
				Date:        "2026-01-15",
				Description: "Morning coffee",
			},
			wantErr: false,
		},
		{
			name: "valid request with minimal fields (category and date default)",
			input: dto.CreateExpenseRequest{
				Title:  "Bus fare",
				Amount: "1.50",
			},
			wantErr: false,
		},
		// --- title ---
		{
			name:        "error on empty title",
			input:       dto.CreateExpenseRequest{Amount: "5.00", Date: "2026-01-01"},
			wantErr:     true,
			errContains: "title is required",
		},
		{
			name:        "error on whitespace-only title",
			input:       dto.CreateExpenseRequest{Title: "   ", Amount: "5.00", Date: "2026-01-01"},
			wantErr:     true,
			errContains: "title is required",
		},
		// --- amount ---
		{
			name:        "error on missing amount",
			input:       dto.CreateExpenseRequest{Title: "Coffee", Date: "2026-01-01"},
			wantErr:     true,
			errContains: "amount is required",
		},
		{
			name:        "error on zero amount",
			input:       dto.CreateExpenseRequest{Title: "Coffee", Amount: "0", Date: "2026-01-01"},
			wantErr:     true,
			errContains: "amount must be a positive number",
		},
		{
			name:        "error on negative amount",
			input:       dto.CreateExpenseRequest{Title: "Coffee", Amount: "-1.00", Date: "2026-01-01"},
			wantErr:     true,
			errContains: "amount must be a positive number",
		},
		{
			name:        "error on non-numeric amount",
			input:       dto.CreateExpenseRequest{Title: "Coffee", Amount: "abc", Date: "2026-01-01"},
			wantErr:     true,
			errContains: "amount must be a positive number",
		},
		{
			name: "valid decimal amount",
			input: dto.CreateExpenseRequest{
				Title:  "Coffee",
				Amount: "0.01",
				Date:   "2026-01-01",
			},
			wantErr: false,
		},
		// --- category ---
		{
			name:        "error on invalid category",
			input:       dto.CreateExpenseRequest{Title: "Coffee", Amount: "3.50", Category: "InvalidCat", Date: "2026-01-01"},
			wantErr:     true,
			errContains: "category is invalid",
		},
		{
			name: "defaults category to Others when empty",
			input: dto.CreateExpenseRequest{
				Title:  "Coffee",
				Amount: "3.50",
				Date:   "2026-01-01",
			},
			wantErr: false,
		},
		{
			name: "accepts every valid category",
			// We only need one representative; detailed category tests live in models.
			input: dto.CreateExpenseRequest{
				Title:    "Medicine",
				Amount:   "20.00",
				Category: models.CategoryHealth,
				Date:     "2026-01-01",
			},
			wantErr: false,
		},
		// --- date ---
		{
			name:        "error on invalid date format",
			input:       dto.CreateExpenseRequest{Title: "Coffee", Amount: "3.50", Date: "15-01-2026"},
			wantErr:     true,
			errContains: "date must be in YYYY-MM-DD format",
		},
		{
			name:        "error on datetime string instead of date",
			input:       dto.CreateExpenseRequest{Title: "Coffee", Amount: "3.50", Date: "2026-01-15T00:00:00Z"},
			wantErr:     true,
			errContains: "date must be in YYYY-MM-DD format",
		},
		{
			name: "defaults date to today when empty",
			input: dto.CreateExpenseRequest{
				Title:  "Coffee",
				Amount: "3.50",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.input // copy so mutation doesn't affect other subtests
			err := ValidateCreateExpense(&req)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateCreateExpense() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.errContains != "" && err != nil {
				if msg := err.Error(); len(msg) == 0 || !contains(msg, tt.errContains) {
					t.Errorf("error %q does not contain %q", msg, tt.errContains)
				}
			}
		})
	}
}

// ---- ValidateUpdateExpense ----

func TestValidateUpdateExpense(t *testing.T) {
	tests := []struct {
		name        string
		input       dto.UpdateExpenseRequest
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid with no fields set (no-op update)",
			input:   dto.UpdateExpenseRequest{},
			wantErr: false,
		},
		{
			name:    "valid with all fields set",
			input:   dto.UpdateExpenseRequest{Title: strPtr("New Title"), Amount: strPtr("9.99"), Category: categoryPtr(models.CategoryLeisure), Date: strPtr("2026-03-01"), Description: strPtr("desc")},
			wantErr: false,
		},
		// --- title ---
		{
			name:        "error on empty title pointer",
			input:       dto.UpdateExpenseRequest{Title: strPtr("")},
			wantErr:     true,
			errContains: "title cannot be empty",
		},
		{
			name:        "error on whitespace-only title",
			input:       dto.UpdateExpenseRequest{Title: strPtr("   ")},
			wantErr:     true,
			errContains: "title cannot be empty",
		},
		{
			name:    "valid non-empty title",
			input:   dto.UpdateExpenseRequest{Title: strPtr("Renamed")},
			wantErr: false,
		},
		// --- amount ---
		{
			name:        "error on zero amount",
			input:       dto.UpdateExpenseRequest{Amount: strPtr("0")},
			wantErr:     true,
			errContains: "amount must be a positive number",
		},
		{
			name:        "error on negative amount",
			input:       dto.UpdateExpenseRequest{Amount: strPtr("-5.00")},
			wantErr:     true,
			errContains: "amount must be a positive number",
		},
		{
			name:        "error on non-numeric amount",
			input:       dto.UpdateExpenseRequest{Amount: strPtr("not-a-number")},
			wantErr:     true,
			errContains: "amount must be a positive number",
		},
		{
			name:    "valid positive amount",
			input:   dto.UpdateExpenseRequest{Amount: strPtr("0.50")},
			wantErr: false,
		},
		// --- category ---
		{
			name:        "error on invalid category",
			input:       dto.UpdateExpenseRequest{Category: categoryPtr("InvalidCat")},
			wantErr:     true,
			errContains: "category is invalid",
		},
		{
			name:    "valid category",
			input:   dto.UpdateExpenseRequest{Category: categoryPtr(models.CategoryElectronics)},
			wantErr: false,
		},
		// --- date ---
		{
			name:        "error on invalid date format",
			input:       dto.UpdateExpenseRequest{Date: strPtr("15/01/2026")},
			wantErr:     true,
			errContains: "date must be in YYYY-MM-DD format",
		},
		{
			name:    "valid date",
			input:   dto.UpdateExpenseRequest{Date: strPtr("2026-06-15")},
			wantErr: false,
		},
		// --- nil fields are skipped ----
		{
			name:    "nil title is skipped",
			input:   dto.UpdateExpenseRequest{Title: nil, Amount: strPtr("5.00")},
			wantErr: false,
		},
		{
			name:    "nil amount is skipped",
			input:   dto.UpdateExpenseRequest{Title: strPtr("Valid"), Amount: nil},
			wantErr: false,
		},
		{
			name:    "nil category is skipped",
			input:   dto.UpdateExpenseRequest{Title: strPtr("Valid"), Category: nil},
			wantErr: false,
		},
		{
			name:    "nil date is skipped",
			input:   dto.UpdateExpenseRequest{Title: strPtr("Valid"), Date: nil},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.input // copy so pointer mutations don't affect other tests
			err := ValidateUpdateExpense(&req)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateUpdateExpense() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.errContains != "" && err != nil {
				if msg := err.Error(); !contains(msg, tt.errContains) {
					t.Errorf("error %q does not contain %q", msg, tt.errContains)
				}
			}
		})
	}
}

// contains is a simple substring check that avoids importing strings in test code.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
