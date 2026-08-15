package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/xenptr/go-projects/todo-list-api/internal/auth"
	"github.com/xenptr/go-projects/todo-list-api/internal/dto"
	"github.com/xenptr/go-projects/todo-list-api/internal/models"
	"github.com/xenptr/go-projects/todo-list-api/internal/validation"
)

func (h *Handler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	defer r.Body.Close()

	var input dto.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validation.CreateTodoRequest(input); err != nil {
		writeValidationError(w, err)
		return
	}

	todo := models.Todo{
		UserID:      userID,
		Title:       input.Title,
		Description: input.Description,
	}

	id, err := h.todoRepo.CreateTodo(r.Context(), todo)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	created, err := h.todoRepo.GetTodoByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ownerID, err := h.todoRepo.GetTodoOwner(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "todo not found")
		return
	}
	if ownerID != userID {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

	defer r.Body.Close()

	var input dto.UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validation.UpdateTodoRequest(input); err != nil {
		writeValidationError(w, err)
		return
	}

	todo := models.Todo{
		UserID:      userID,
		Title:       input.Title,
		Description: input.Description,
		Completed:   input.Completed,
	}

	if err = h.todoRepo.UpdateTodo(r.Context(), id, todo); err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	updated, err := h.todoRepo.GetTodoByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ownerID, err := h.todoRepo.GetTodoOwner(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "todo not found")
		return
	}
	if ownerID != userID {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

	if err = h.todoRepo.DeleteTodo(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListTodos(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	values := r.URL.Query()

	var page *int64
	var limit *int64
	var search *string

	if value := values.Get("page"); value != "" {
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil || v < 1 {
			writeError(w, http.StatusBadRequest, "page must be a positive integer")
			return
		}
		page = &v
	}

	if value := values.Get("limit"); value != "" {
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil || v < 1 {
			writeError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		limit = &v
	}

	if value := values.Get("search"); value != "" {
		search = &value
	} else if value := values.Get("s"); value != "" {
		search = &value
	}

	if page != nil && limit == nil {
		writeError(w, http.StatusBadRequest, "limit is required when page is provided")
		return
	}

	todos, total, err := h.todoRepo.ListTodosByUser(r.Context(), userID, page, limit, search)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response := map[string]any{
		"data":  todos,
		"total": total,
	}
	if page != nil {
		response["page"] = *page
	}
	if limit != nil {
		response["limit"] = *limit
	}

	writeJSON(w, http.StatusOK, response)
}
