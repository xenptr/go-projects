package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/xenptr/go-projects/todo-list-api/internal/auth"
	"github.com/xenptr/go-projects/todo-list-api/internal/models"
)

func (h *Handler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userToken := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(userToken, "Bearer ")

	userID, err := auth.ParseToken(tokenString, h.secret)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	var todo models.Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	todo.UserID = userID

	id, err := h.repo.CreateTodo(todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	created, err := h.repo.GetTodoByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}
