package models

import "time"

type Todo struct {
	ID          int64
	UserID      int64
	Title       string
	Description string
	Completed   bool
	CreatedAt   time.Time
}
