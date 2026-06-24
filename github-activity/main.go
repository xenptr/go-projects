package main

import (
	"fmt"
	"os"
)

func main() {
	var username string

	if len(os.Args) < 2 {
		fmt.Println("Usage: ./github-activity <username>")
		os.Exit(1)
	}

	username = os.Args[1]

	events, err := fetchEvents(username)
	if err != nil {
		panic(err)
	}

	if len(events) == 0 {
		fmt.Printf("No recent public activity found for %s\n", username)
		return
	}

	for _, event := range events {
		if msg := formatEvent(event); msg != "" {
			fmt.Print(msg)
		}
	}

	// enc := json.NewEncoder(os.Stdout)
	// enc.SetIndent("", "  ")
	// enc.Encode(events)
}
