package main

import "time"

type Blog struct {
	Articles []Article
}

type Article struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	PublishedAt time.Time `json:"published_at"`
	Content     string    `json:"content"`
	Category    string    `json:"category"`
	Tags        []string  `json:"tags"`
	Comments    []Comment `json:"comments"`
	Preview     string    `json:"-"`
}

type Comment struct {
	ID        int       `json:"id"`
	ArticleID int       `json:"article_id"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type ArticlePageData struct {
	*Article
	CommentError string
}
