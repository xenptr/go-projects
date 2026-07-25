package store

import "github.com/xenptr/go-projects/blogging-platform-api/internal/models"

type PostStore interface {
	AddPost(post models.Post) (int64, error)
	AllPosts() ([]models.Post, error)
	PostByID(id int64) (models.Post, error)
	PostsByTerm(term string) ([]models.Post, error)
	UpdatePost(id int64, post models.Post) error
	DeletePost(id int64) error
}
