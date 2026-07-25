package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/xenptr/go-projects/blogging-platform-api/internal/models"
	"github.com/xenptr/go-projects/blogging-platform-api/internal/store"
)

// ── response helpers

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

// errorResponse is the standard JSON error envelope returned to callers.
type errorResponse struct {
	Error string `json:"error"`
}

// writeError writes a JSON error envelope with the given status code.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// storeError maps store sentinel errors to the appropriate HTTP status and
// response body. Any error that is not a recognised sentinel becomes a 500.
func storeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, store.ErrInvalidInput):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, store.ErrNoUpdate):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

// parseID extracts and validates the {id} path value.
func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(r.PathValue("id")), 10, 64)
}

// ── handlers

func (h *Handler) ListPosts(w http.ResponseWriter, r *http.Request) {
	var (
		posts []models.Post
		err   error
	)

	term := strings.TrimSpace(r.URL.Query().Get("term"))
	if term != "" {
		posts, err = h.store.PostsByTerm(term)
	} else {
		posts, err = h.store.AllPosts()
	}
	if err != nil {
		storeError(w, err)
		return
	}

	if posts == nil {
		posts = []models.Post{}
	}

	writeJSON(w, http.StatusOK, posts)
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var post models.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		writeError(w, http.StatusBadRequest, "malformed request body")
		return
	}

	id, err := h.store.AddPost(post)
	if err != nil {
		storeError(w, err)
		return
	}

	created, err := h.store.PostByID(id)
	if err != nil {
		storeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	post, err := h.store.PostByID(id)
	if err != nil {
		storeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, post)
}

func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	defer r.Body.Close()

	var post models.Post
	if err = json.NewDecoder(r.Body).Decode(&post); err != nil {
		writeError(w, http.StatusBadRequest, "malformed request body")
		return
	}

	if err = h.store.UpdatePost(id, post); err != nil {
		storeError(w, err)
		return
	}

	updated, err := h.store.PostByID(id)
	if err != nil {
		storeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	if err = h.store.DeletePost(id); err != nil {
		storeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
