package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	dataDir    = "data"
	articleDir = "data/articles"
	nextIDFile = "data/next_id.txt"
)

var nextId int

func init() {
	if err := os.MkdirAll(articleDir, 0o755); err != nil {
		log.Fatal(err)
	}

	var err error
	nextId, err = loadNextID()
	if err != nil {
		log.Fatal(err)
	}
}

func loadNextID() (int, error) {
	var data []byte
	if _, err := os.Stat(nextIDFile); err == nil {
		data, err = os.ReadFile(nextIDFile)
		if err != nil {
			return 0, err
		}
		return strconv.Atoi(string(data))
	}

	entries, err := os.ReadDir(articleDir)
	if err != nil {
		return 0, err
	}

	maxID := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		idStr := strings.TrimSuffix(name, filepath.Ext(name))

		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		if id > maxID {
			maxID = id
		}
	}

	return maxID + 1, nil
}

func saveNextID(id int) error {
	return os.WriteFile(nextIDFile, []byte(strconv.Itoa(id)), 0o644)
}

func formatDate(date time.Time) string {
	return date.Format("02 Jan 2006")
}

func inputDate(date time.Time) string {
	return date.Format("2006-01-02")
}

type Blog struct {
	Articles []Article
}

type Article struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	PublishedAt time.Time `json:"published_at"`
	Content     string    `json:"content"`
}

func (a *Article) save() error {
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(articleDir, fmt.Sprintf("%d.json", a.ID))
	return os.WriteFile(filename, data, 0o644)
}

func loadBlog() (*Blog, error) {
	entries, err := os.ReadDir(articleDir)
	if err != nil {
		return nil, err
	}

	blog := &Blog{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := filepath.Join(articleDir, entry.Name())
		data, err := os.ReadFile(filename)
		if err != nil {
			continue
		}

		var article Article

		if err = json.Unmarshal(data, &article); err != nil {
			log.Printf("failed to load %s: %v", filename, err)
			continue
		}

		blog.Articles = append(blog.Articles, article)
	}

	sort.Slice(blog.Articles, func(i, j int) bool {
		return blog.Articles[i].PublishedAt.After(blog.Articles[j].PublishedAt)
	})

	return blog, nil
}

func loadArticle(id int) (*Article, error) {
	filename := filepath.Join(articleDir, fmt.Sprintf("%d.json", id))

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	article := &Article{}

	if err = json.Unmarshal(data, &article); err != nil {
		return nil, err
	}

	return article, nil
}

func deleteArticle(id int) error {
	filename := filepath.Join(articleDir, fmt.Sprintf("%d.json", id))
	return os.Remove(filename)
}

func parsePublishedDate(r *http.Request) (time.Time, error) {
	return time.Parse("2006-01-02", r.FormValue("published_at"))
}

var templates = map[string]*template.Template{}

func loadTemplates() error {
	entries, err := os.ReadDir("templates")
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"formatDate": formatDate,
		"inputDate":  inputDate,
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if name == "layout.html" {
			continue
		}

		page := strings.TrimSuffix(name, filepath.Ext(name))

		templates[page] = template.Must(template.New("").
			Funcs(funcMap).
			ParseFiles(
				"templates/layout.html",
				filepath.Join("templates", name),
			),
		)

	}

	return nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	t, ok := templates[tmpl]
	if !ok {
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}

	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	blog, err := loadBlog()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "home", blog)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	article, err := loadArticle(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	renderTemplate(w, "article", article)
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	blog, err := loadBlog()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

		publishedAt, err := parsePublishedDate(r)
		if err != nil {
			http.Error(w, "Invalid date", http.StatusBadRequest)
			return
		}

		article := &Article{
			ID:          nextId,
			Title:       r.FormValue("title"),
			PublishedAt: publishedAt,
			Content:     r.FormValue("content"),
		}

		if err = article.save(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		nextId++
		if err := saveNextID(nextId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/article/%d", article.ID), http.StatusSeeOther)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	article, err := loadArticle(id)
	if err != nil {
		http.NotFound(w, r)
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

		publishedAt, err := parsePublishedDate(r)
		if err != nil {
			http.Error(w, "Invalid date", http.StatusBadRequest)
			return
		}

		article.Title = r.FormValue("title")
		article.PublishedAt = publishedAt
		article.Content = r.FormValue("content")

		if err = article.save(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := deleteArticle(id); err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func main() {
	if err := loadTemplates(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/article/{id}", viewHandler)

	http.HandleFunc("/admin", adminHandler)
	http.HandleFunc("/admin/new", newHandler)
	http.HandleFunc("/admin/edit/{id}", editHandler)
	http.HandleFunc("/admin/delete/{id}", deleteHandler)

	http.Handle("/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.Dir("static")),
		),
	)

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
