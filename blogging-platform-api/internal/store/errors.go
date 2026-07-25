package store

import "errors"

// Sentinel errors returned by all store implementations.
var (
	// ErrNotFound is returned when a requested record does not exist.
	ErrNotFound = errors.New("post not found")

	// ErrInvalidInput is returned when the caller supplies invalid or
	// incomplete data (e.g. missing required fields).
	ErrInvalidInput = errors.New("invalid input")

	// ErrNoUpdate is returned by UpdatePost when no updatable fields
	// were present in the payload.
	ErrNoUpdate = errors.New("nothing to update")
)
