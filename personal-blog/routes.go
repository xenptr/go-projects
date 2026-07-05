package main

import "net/http"

func registerRoutes() {
	// Public
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/article/{id}", viewHandler)
	http.HandleFunc("/article/{id}/comment", commentHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/category/{name}", categoryHandler)
	http.HandleFunc("/tag/{name}", tagHandler)

	// Admin
	http.HandleFunc("/admin", requireAuth(adminHandler))
	http.HandleFunc("/admin/new", requireAuth(newHandler))
	http.HandleFunc("/admin/edit/{id}", requireAuth(editHandler))
	http.HandleFunc("/admin/delete/{id}", requireAuth(deleteHandler))

	// Static assets
	http.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.Dir("static")),
		),
	)
}
