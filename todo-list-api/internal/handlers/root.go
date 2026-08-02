package handlers

import (
	"fmt"
	"net/http"
)

func (h *Handler) Root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to Todo List API!")
}
