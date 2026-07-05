package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	blog, err := loadBlog()
	if err != nil {
		render500(w, err)
		return
	}
	renderTemplate(w, "home", blog)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	id, err := articleID(r)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	article, err := loadArticle(id)
	if err != nil {
		render404(w)
		return
	}
	renderTemplate(w, "article", &ArticlePageData{Article: article})
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	blog, err := loadBlog()
	if err != nil {
		render500(w, err)
		return
	}

	data := struct {
		Query    string
		Articles []Article
		Blog     *Blog
	}{
		Query: query,
		Blog:  blog,
	}

	if query != "" {
		data.Articles = searchArticles(blog, query)
	}

	renderTemplate(w, "search", data)
}

func categoryHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	blog, err := loadBlog()
	if err != nil {
		render500(w, err)
		return
	}

	data := struct {
		Filter   string
		Articles []Article
		Blog     *Blog
	}{
		Filter:   name,
		Articles: filterByCategory(blog, name),
		Blog:     blog,
	}

	renderTemplate(w, "category", data)
}

func tagHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	blog, err := loadBlog()
	if err != nil {
		render500(w, err)
		return
	}

	data := struct {
		Filter   string
		Articles []Article
		Blog     *Blog
	}{
		Filter:   name,
		Articles: filterByTag(blog, name),
		Blog:     blog,
	}

	renderTemplate(w, "tag", data)
}

func commentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := articleID(r)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	author := strings.TrimSpace(r.FormValue("author"))
	body := strings.TrimSpace(r.FormValue("body"))

	var commentErr string
	switch {
	case author == "":
		commentErr = "Name is required."
	case len(author) > 100:
		commentErr = "Name is too long."
	case body == "":
		commentErr = "Comment body is required."
	case len(body) > 2000:
		commentErr = "Comment is too long (max 2000 characters)."
	}

	if commentErr != "" {
		article, err := loadArticle(id)
		if err != nil {
			render404(w)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		renderTemplate(w, "article", &ArticlePageData{Article: article, CommentError: commentErr})
		return
	}

	if err := addComment(id, author, body); err != nil {
		render500(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/article/%d#comments", id), http.StatusSeeOther)
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	blog, err := loadBlog()
	if err != nil {
		render500(w, err)
		return
	}
	renderTemplate(w, "dashboard", blog)
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		renderTemplate(w, "new", nil)

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		title := strings.TrimSpace(r.FormValue("title"))
		content := strings.TrimSpace(r.FormValue("content"))
		category := strings.TrimSpace(r.FormValue("category"))
		tags := splitTags(r.FormValue("tags"))

		publishedAt, err := parsePublishedDate(r)
		if err != nil {
			http.Error(w, "Invalid date", http.StatusBadRequest)
			return
		}

		if err := validateArticle(title, content); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		article := &Article{
			ID:          nextID,
			Title:       title,
			PublishedAt: publishedAt,
			Content:     content,
			Category:    category,
			Tags:        tags,
		}

		if err = article.save(); err != nil {
			render500(w, err)
			return
		}

		nextID++
		if err := saveNextID(nextID); err != nil {
			render500(w, err)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/article/%d", article.ID), http.StatusSeeOther)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	id, err := articleID(r)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	article, err := loadArticle(id)
	if err != nil {
		render404(w)
		return
	}

	switch r.Method {
	case http.MethodGet:
		renderTemplate(w, "edit", article)

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		title := strings.TrimSpace(r.FormValue("title"))
		content := strings.TrimSpace(r.FormValue("content"))
		category := strings.TrimSpace(r.FormValue("category"))
		tags := splitTags(r.FormValue("tags"))

		publishedAt, err := parsePublishedDate(r)
		if err != nil {
			http.Error(w, "Invalid date", http.StatusBadRequest)
			return
		}

		if err := validateArticle(title, content); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		article.Title = title
		article.PublishedAt = publishedAt
		article.Content = content
		article.Category = category
		article.Tags = tags

		if err = article.save(); err != nil {
			render500(w, err)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/article/%d", article.ID), http.StatusSeeOther)
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := articleID(r)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := deleteArticle(id); err != nil {
		if os.IsNotExist(err) {
			render404(w)
			return
		}
		render500(w, err)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}