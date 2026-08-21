package repository

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrDuplicateEmail = errors.New("email already in use")
)
