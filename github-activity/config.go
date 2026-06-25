package main

import (
	"flag"
	"os"
)

type Config struct {
	Username string
	Filter   string
	Types    bool
	Limit    int
	JSON     bool
}

func parseFlags() Config {
	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	filter := flag.String("filter", "", "Filter by event type (e.g. PushEvent)")
	types := flag.Bool("types", false, "List all supported event types")
	limit := flag.Int("limit", 0, "Limit the number of events displayed (0 = no limit, max 100 per GitHub API)")
	jsonOutput := flag.Bool("json", false, "Output raw JSON")

	if err := flag.CommandLine.Parse(os.Args[2:]); err != nil {
		os.Exit(2)
	}

	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(1)
	}

	return Config{
		Username: os.Args[1],
		Filter:   *filter,
		Types:    *types,
		Limit:    *limit,
		JSON:     *jsonOutput,
	}

}
