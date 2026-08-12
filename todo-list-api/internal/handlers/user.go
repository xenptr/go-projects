package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/xenptr/go-projects/todo-list-api/internal/auth"
	"github.com/xenptr/go-projects/todo-list-api/internal/models"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	var buf bytes.Buffer

	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hash, err := auth.HashPassword(user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user.PasswordHash = hash
	user.Password = ""

	id, err := h.repo.CreateUser(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	signed, err := auth.GenerateToken(id, h.secret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	created, err := h.repo.GetUserByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"user":  created,
		"token": signed,
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// userToken := r.Header.Get("Authorization")
	// t := strings.TrimPrefix(userToken, "Bearer ")
	// claims := jwt.RegisteredClaims{
	// 	Subject:   strconv.FormatInt(user.ID, 10),
	// 	ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	// 	IssuedAt:  jwt.NewNumericDate(time.Now()),
	// }
	// parser := jwt.NewParser(
	// 	jwt.WithValidMethods([]string{"HS256"}),
	// )
	// keyFunc := func(token *jwt.Token) (any, error) {
	// 	return h.secret, nil
	// }
	// token, err := parser.ParseWithClaims(t, &claims, keyFunc)
	// if err != nil {
	// 	if errors.Is(err, jwt.ErrTokenExpired) {
	// 		http.Error(w, "token expired", http.StatusUnauthorized)
	// 		return
	// 	}
	//
	// 	if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
	// 		http.Error(w, "invalid token", http.StatusUnauthorized)
	// 		return
	// 	}
	//
	// 	http.Error(w, "invalid token", http.StatusUnauthorized)
	// 	return
	// }

	data, err := h.repo.GetUserByEmail(user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = auth.CheckPassword(user.Password, data.PasswordHash); err != nil {
		http.Error(w, "inavlid credentials", http.StatusUnauthorized)
		return
	}

	signed, err := auth.GenerateToken(data.ID, h.secret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}

	response := map[string]any{
		"user":  data,
		"token": signed,
	}

	writeJSON(w, http.StatusOK, response)
}
