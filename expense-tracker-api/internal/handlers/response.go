package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/xenptr/go-projects/expense-tracker-api/internal/validation"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, _ = w.Write(buf.Bytes())
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"message": message})
}

func writeValidationError(w http.ResponseWriter, err error) {
	ve, ok := err.(*validation.ValidationError)
	if !ok {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusBadRequest, map[string]any{
		"message": "validation failed",
		"errors":  ve.Errors,
	})
}
