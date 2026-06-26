package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

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

	if cfg.Profile {
		var (
			data User
			err  error
		)

		if cfg.Cache {
			data, err = fetchUserWithCache(cfg.Username)
		} else {
			data, err = fetchUser(cfg.Username)
		}
		if err != nil {
			fatal(err)
		}

		if cfg.JSON {
			if err := printJSON(data); err != nil {
				fatal(err)
			}
			return
		}

		fmt.Print(formatUser(data))
	}

	var (
		events []Event
		err    error
	)

	if cfg.Cache {
		events, err = fetchUserEventsWithCache(cfg.Username, cfg.Limit)
	} else {
		events, err = fetchUserEvents(cfg.Username, cfg.Limit)
	}
	if err != nil {
		fatal(err)
	}

	if cfg.JSON {
		if err := printJSON(events); err != nil {
			fatal(err)
		}
		return
	}

	if len(events) == 0 {
		fmt.Printf("No recent public activity found for %s\n", cfg.Username)
		return
	}

	fmt.Printf("Recent activity for %s:\n\n", cfg.Username)
	for _, event := range events {
		if filterType != "" && event.Type != filterType {
			continue
		}

		if msg := formatEvent(event); msg != "" {
			fmt.Print("  ", msg)
		}
	}
}
