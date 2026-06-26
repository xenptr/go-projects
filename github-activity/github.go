package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	apiBase    = "https://api.github.com"
	apiVersion = "2026-03-10"

	eventsEndpoint = "/users/%s/events"
	userEndpoint   = "/users/%s"
)

var client = &http.Client{}

func githubRequest(url string, out any) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", apiVersion)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// ok
	case http.StatusNotFound:
		return fmt.Errorf("user not found")
	case http.StatusForbidden, http.StatusTooManyRequests:
		return fmt.Errorf("GitHub API rate limit exceeded, try again later")
	default:
		return fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return err
	}
	return nil
}

func githubGetJSON[T any](url string) (T, error) {
	var data T
	err := githubRequest(url, &data)
	return data, err
}

func userEventsURL(username string, limit int) string {
	base := fmt.Sprintf(apiBase+eventsEndpoint, username)
	if limit > 0 {
		// GitHub API caps per_page at 100
		if limit > 100 {
			limit = 100
		}
		return fmt.Sprintf("%s?per_page=%d", base, limit)
	}
	return base
}

func fetchUserEvents(username string, limit int) ([]Event, error) {
	return githubGetJSON[[]Event](userEventsURL(username, limit))
}

// func fetchRawUserEvents(username string, limit int) ([]any, error) {
// 	return githubGetJSON[[]any](userEventsURL(username, limit))
// }

func userURL(username string) string {
	return fmt.Sprintf(apiBase+userEndpoint, username)
}

func fetchUser(username string) (User, error) {
	return githubGetJSON[User](userURL(username))
}

// func fetchRawUser(username string) (any, error) {
// 	return githubGetJSON[any](userURL(username))
// }
