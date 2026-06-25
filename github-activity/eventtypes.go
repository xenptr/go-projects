package main

import (
	"fmt"
	"slices"
	"strings"
)

func printEventTypes() {
	fmt.Println("Available types")
	for _, t := range EventTypes {
		fmt.Print("  ")
		fmt.Println(t)
	}
}

func parseEventType(s string) (EventType, bool) {
	s = strings.TrimSpace(s)

	if s == "" {
		return "", true
	}

	eventType := EventType(s)

	if !slices.Contains(EventTypes, eventType) {
		return "", false
	}

	return eventType, true
}
