package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	apiBase        = "https://api.github.com"
	apiVersion     = "2026-03-10"
	eventsEndpoint = "/users/%s/events"
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github API returned %s", resp.Status)
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

func fetchRawUserEvents(username string, limit int) ([]any, error) {
	return githubGetJSON[[]any](userEventsURL(username, limit))
}
