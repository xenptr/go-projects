package models

import "time"

type Todo struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"-"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"copmleted"`
	CreatedAt   time.Time `json:"created_at"`
}
