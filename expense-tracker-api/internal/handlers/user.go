package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/xenptr/go-projects/expense-tracker-api/internal/auth"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/dto"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/models"
	"github.com/xenptr/go-projects/expense-tracker-api/internal/validation"
)

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var input dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validation.RegisterRequest(input); err != nil {
		writeValidationError(w, err)
		return
	}

	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	user := models.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: hash,
	}

	id, err := h.userRepo.CreateUser(r.Context(), user)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			writeError(w, http.StatusConflict, "email already in use")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp, err := h.issueTokenPair(r, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var input dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validation.LoginRequest(input); err != nil {
		writeValidationError(w, err)
		return
	}

	user, err := h.userRepo.GetUserByEmail(r.Context(), input.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err = auth.CheckPassword(input.Password, user.PasswordHash); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	resp, err := h.issueTokenPair(r, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// RefreshToken validates a refresh token, revokes it (rotation), and issues a
// fresh access + refresh token pair. This prevents refresh token reuse attacks.
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var input dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validation.RefreshRequest(input); err != nil {
		writeValidationError(w, err)
		return
	}

	// Validate the JWT signature and expiry.
	userID, err := auth.ParseRefreshToken(input.RefreshToken, h.secret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	// Check the token is still in the store (not revoked).
	exists, err := h.refreshStore.Exists(r.Context(), userID, input.RefreshToken)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		writeError(w, http.StatusUnauthorized, "refresh token has been revoked")
		return
	}

	// Verify the user account still exists.
	if _, err = h.userRepo.GetUserByID(r.Context(), userID); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	// Revoke the used refresh token before issuing a new one (token rotation).
	if err = h.refreshStore.Revoke(r.Context(), userID, input.RefreshToken); err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp, err := h.issueTokenPair(r, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// issueTokenPair generates a new access + refresh token pair, persists the
// refresh token in the store, and returns the response DTO.
func (h *Handler) issueTokenPair(r *http.Request, userID int64) (dto.AuthResponse, error) {
	token, err := auth.GenerateToken(userID, h.secret)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	refreshToken, err := auth.GenerateRefreshToken(userID, h.secret)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	if err = h.refreshStore.Save(r.Context(), userID, refreshToken, auth.RefreshTokenTTL); err != nil {
		return dto.AuthResponse{}, err
	}

	return dto.AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}
