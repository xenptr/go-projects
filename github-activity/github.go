package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const apiBase = "https://api.github.com"

func githubGet(url string, out any) error {
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")

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

func fetchEvents(username string) ([]Event, error) {
	url := fmt.Sprintf("%s/users/%s/events", apiBase, username)

	var events []Event
	if err := githubGet(url, &events); err != nil {
		return nil, err
	}

	return events, nil
}

