package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func main() {
	var username string

	if len(os.Args) < 2 {
		fmt.Println("Usage: ./github-activity <username>")
		os.Exit(1)
	}

	username = os.Args[1]

	githubApi := fmt.Sprintf("https://api.github.com/users/%s/events", username)
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", githubApi, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept","application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Github API returned %s\n", resp.Status)
		os.Exit(1)
	}

	var data any

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		panic(err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(data)
}
