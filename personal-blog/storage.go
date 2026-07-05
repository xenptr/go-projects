package main

import (
	"encoding/json"
	"fmt"
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
	articleDir = "data/articles"
	nextIDFile = "data/next_id.txt"
)

var nextID int

func init() {
	if err := os.MkdirAll(articleDir, 0o755); err != nil {
		log.Fatal(err)
	}

	var err error
	nextID, err = loadNextID()
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
		return strconv.Atoi(strings.TrimSpace(string(data)))
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

func (a *Article) save() error {
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(articleDir, fmt.Sprintf("%d.json", a.ID))
	return os.WriteFile(filename, data, 0o644)
}

// loadBlog loads all articles sorted newest-first.
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

		article.Preview = contentPreview(article.Content, 160)
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

func addComment(articleID int, author, body string) error {
	article, err := loadArticle(articleID)
	if err != nil {
		return err
	}

	// Determine next comment ID within the article.
	maxCID := 0
	for _, c := range article.Comments {
		if c.ID > maxCID {
			maxCID = c.ID
		}
	}

	comment := Comment{
		ID:        maxCID + 1,
		ArticleID: articleID,
		Author:    author,
		Body:      body,
		CreatedAt: time.Now().UTC(),
	}

	article.Comments = append(article.Comments, comment)
	return article.save()
}

func searchArticles(blog *Blog, query string) []Article {
	q := strings.ToLower(query)
	var results []Article
	for _, a := range blog.Articles {
		if strings.Contains(strings.ToLower(a.Title), q) ||
			strings.Contains(strings.ToLower(a.Content), q) {
			results = append(results, a)
		}
	}
	return results
}

func filterByCategory(blog *Blog, category string) []Article {
	var results []Article
	for _, a := range blog.Articles {
		if strings.EqualFold(a.Category, category) {
			results = append(results, a)
		}
	}
	return results
}

func filterByTag(blog *Blog, tag string) []Article {
	var results []Article
	for _, a := range blog.Articles {
		for _, t := range a.Tags {
			if strings.EqualFold(t, tag) {
				results = append(results, a)
				break
			}
		}
	}
	return results
}

func articleID(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}
