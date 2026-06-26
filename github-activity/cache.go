package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const cacheTTL = 5 * time.Minute

type cacheEntry[T any] struct {
	FetchedAt time.Time `json:"fetched_at"`
	Data      T         `json:"data"`
}

func cacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(base, "github-activity")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}

	return dir, nil
}

func cacheKey(key string) string {
	return fmt.Sprintf("%x.json", md5.Sum([]byte(key)))
}

func cacheLoad[T any](key string) (T, bool) {
	var zero T

	dir, err := cacheDir()
	if err != nil {
		return zero, false
	}

	path := filepath.Join(dir, cacheKey(key))
	data, err := os.ReadFile(path)
	if err != nil {
		return zero, false
	}

	var entry cacheEntry[T]
	if err := json.Unmarshal(data, &entry); err != nil {
		return zero, false
	}

	if time.Since(entry.FetchedAt) > cacheTTL {
		return zero, false
	}

	return entry.Data, true
}

func cacheSave[T any](key string, value T) {
	dir, err := cacheDir()
	if err != nil {
		return
	}

	entry := cacheEntry[T]{
		FetchedAt: time.Now(),
		Data:      value,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	path := filepath.Join(dir, cacheKey(key))
	_ = os.WriteFile(path, data, 0o600)
}

func fetchUserEventsWithCache(username string, limit int) ([]Event, error) {
	key := fmt.Sprintf("events:%s:%d", username, limit)

	if cached, ok := cacheLoad[[]Event](key); ok {
		return cached, nil
	}

	events, err := fetchUserEvents(username, limit)
	if err != nil {
		return nil, err
	}

	cacheSave(key, events)

	return events, nil
}

// func fetchRawUserEventsWithCache(username string, limit int) ([]any, error) {
// 	key := fmt.Sprintf("raw_events:%s:%d", username, limit)

// 	if cached, ok := cacheLoad[[]any](key); ok {
// 		return cached, nil
// 	}

// 	events, err := fetchRawUserEvents(username, limit)
// 	if err != nil {
// 		return nil, err
// 	}

// 	cacheSave(key, events)

// 	return events, nil
// }

func fetchUserWithCache(username string) (User, error) {
	key := "user:" + username

	if cached, ok := cacheLoad[User](key); ok {
		return cached, nil
	}

	var (
		user User
		err  error
	)

	user, err = fetchUser(username)
	if err != nil {
		return user, err
	}

	cacheSave(key, user)

	return user, nil
}

// func fetchRawUserWithCache(username string) (any, error) {
// 	key := "raw_user:" + username

// 	if cached, ok := cacheLoad[any](key); ok {
// 		return cached, nil
// 	}

// 	var (
// 		user User
// 		err  error
// 	)

// 	user, err = fetchUser(username)
// 	if err != nil {
// 		return user, err
// 	}

// 	cacheSave(key, user)

// 	return user, nil
// }
