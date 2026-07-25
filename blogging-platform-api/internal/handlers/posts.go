package handlers

import (
	"blogging-platform-api/internal/models"
	"blogging-platform-api/internal/store"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) ListPosts(w http.ResponseWriter, r *http.Request) {
	var posts []models.Post
	var err error

	term := strings.TrimSpace(r.URL.Query().Get("term"))
	if term != "" {
		posts, err = h.store.PostsByTerm(term)
	} else {
		posts, err = h.store.AllPosts()
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(posts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CreatePosts(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var post models.Post

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.store.AddBlog(post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	post, err = h.store.PostByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.PathValue("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var post models.Post

	post, err = h.store.PostByID(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.PathValue("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var post models.Post

	defer r.Body.Close()

	if err = json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.store.UpdatePost(id, post)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	post, err = h.store.PostByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.PathValue("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err = h.store.DeletePost(id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
