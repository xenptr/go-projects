package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/auth"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/dto"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/models"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/repository"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/validation"
)

// ListExpenses handles GET /expenses
// Query params: filter=week|month|3months|custom, start_date=YYYY-MM-DD, end_date=YYYY-MM-DD
func (h *Handler) ListExpenses(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query()
	filterStr := q.Get("filter")

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	var filter repository.ExpenseFilter

	switch filterStr {
	case "week":
		start := today.AddDate(0, 0, -7)
		filter.StartDate = &start
	case "month":
		start := today.AddDate(0, -1, 0)
		filter.StartDate = &start
	case "3months":
		start := today.AddDate(0, -3, 0)
		filter.StartDate = &start
	case "custom":
		startStr := q.Get("start_date")
		endStr := q.Get("end_date")
		if startStr == "" || endStr == "" {
			writeError(w, http.StatusBadRequest, "start_date and end_date are required for custom filter")
			return
		}
		start, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "start_date must be YYYY-MM-DD")
			return
		}
		end, err := time.Parse("2006-01-02", endStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "end_date must be YYYY-MM-DD")
			return
		}
		filter.StartDate = &start
		filter.EndDate = &end
	}

	expenses, err := h.expenseRepo.ListExpenses(r.Context(), userID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if expenses == nil {
		expenses = []*models.Expense{}
	}
	writeJSON(w, http.StatusOK, expenses)
}

// CreateExpense handles POST /expenses
func (h *Handler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input dto.CreateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validation.ValidateCreateExpense(&input); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	amount, _ := decimal.NewFromString(input.Amount) // already validated
	date, _ := time.Parse("2006-01-02", input.Date)  // already validated

	expense, err := h.expenseRepo.CreateExpense(r.Context(), userID, input.Title, amount, input.Category, date, input.Description)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, expense)
}

// GetExpense handles GET /expenses/{id}
func (h *Handler) GetExpense(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	expense, err := h.expenseRepo.GetExpense(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "expense not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, expense)
}

// UpdateExpense handles PUT /expenses/{id}
func (h *Handler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	var input dto.UpdateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validation.ValidateUpdateExpense(&input); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	var amount *decimal.Decimal
	if input.Amount != nil {
		a, _ := decimal.NewFromString(*input.Amount) // already validated
		amount = &a
	}

	var date *time.Time
	if input.Date != nil {
		d, _ := time.Parse("2006-01-02", *input.Date) // already validated
		date = &d
	}

	expense, err := h.expenseRepo.UpdateExpense(r.Context(), id, userID, input.Title, amount, input.Category, date, input.Description)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "expense not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, expense)
}

// DeleteExpense handles DELETE /expenses/{id}
func (h *Handler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	if err := h.expenseRepo.DeleteExpense(r.Context(), id, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "expense not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// parseID reads the {id} path value from Go 1.22+ pattern routing.
func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}
