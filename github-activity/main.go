package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: github-activity <username> [options]")
		flag.PrintDefaults()
	}

	cfg := parseFlags()

	if cfg.Types {
		printEventTypes()
		return
	}

	filterType, ok := parseEventType(cfg.Filter)

	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown event type: %s\n\n", cfg.Filter)
		printEventTypes()
		os.Exit(1)
	}

	if cfg.JSON {
		data, err := fetchRawUserEvents(cfg.Username, cfg.Limit)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(data); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	events, err := fetchUserEvents(cfg.Username, cfg.Limit)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(events) == 0 {
		fmt.Printf("No recent public activity found for %s\n", cfg.Username)
		return
	}

	for _, event := range events {
		if filterType != "" && event.Type != filterType {
			continue
		}

		if msg := formatEvent(event); msg != "" {
			fmt.Print(msg)
		}
	}
}
