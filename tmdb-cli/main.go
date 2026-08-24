package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		printUsage(os.Stderr)
		os.Exit(1)
	}

	cfg, err := parseFlags(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printUsage(os.Stdout)
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", err)
		printUsage(os.Stderr)
		os.Exit(1)
	}

	if cfg.ShowVersion {
		fmt.Printf("%s version %s\n", appName, appVersion)
		return
	}

	if cfg.APIKey == "" {
		fmt.Fprintln(os.Stderr, "Error: TMDB API key is missing.")
		fmt.Fprintln(os.Stderr, "Please provide an API key using the --api-key flag or set the TMDB_API_KEY environment variable.")
		fmt.Fprintln(os.Stderr, "\nGet a free API key at: https://www.themoviedb.org/settings/api")
		os.Exit(1)
	}

	client := NewClient(cfg.APIKey)
	ctx := context.Background()

	resp, err := client.FetchMovies(ctx, cfg.MovieType, cfg.Page)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching movies: %s\n", err)
		os.Exit(1)
	}

	if cfg.JSON {
		if cfg.Limit > 0 && cfg.Limit < len(resp.Results) {
			resp.Results = resp.Results[:cfg.Limit]
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(resp); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON output: %s\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Print(formatResponse(resp, cfg.MovieType, cfg.Detailed, cfg.Limit))
}
