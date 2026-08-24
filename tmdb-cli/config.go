package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	appVersion = "1.0.0"
	appName    = "tmdb-app"
)

// Config holds the application configuration parsed from flags and environment variables.
type Config struct {
	Type        string
	MovieType   MovieType
	APIKey      string
	Page        int
	Limit       int
	JSON        bool
	Detailed    bool
	ShowVersion bool
}

// printUsage prints a formatted help message to the given writer.
func printUsage(w io.Writer) {
	fmt.Fprintf(w, `TMDB CLI Tool - Fetch and display movies from The Movie Database

Usage:
  %s --type <type> [options]

Movie Types:
  playing    Now playing movies in theatres
  popular    Current popular movies
  top        Top rated movies of all time
  upcoming   Upcoming movies in theatres

Options:
  --type string      Movie list type (playing, popular, top, upcoming) (required)
  --api-key string   TMDB API Key or Bearer Read Access Token (or set TMDB_API_KEY env var)
  --page int         Page number to fetch (default 1)
  --limit int        Maximum number of results to display (default: all on page)
  --detailed         Show full overview, genres, and additional movie details
  --json             Output raw JSON data
  --version, -v      Show application version
  --help, -h         Show this help message

Examples:
  %s --type "playing"
  %s --type "popular"
  %s --type "top"
  %s --type "upcoming"
  %s --type "popular" --limit 5
  %s --type "top" --page 2 --detailed
  %s --type "playing" --json

Environment Variables:
  TMDB_API_KEY       TMDB API key or Bearer Read Access Token
`, appName, appName, appName, appName, appName, appName, appName, appName)
}

// parseFlags parses command-line arguments and returns a Config.
func parseFlags(args []string) (Config, error) {
	fs := flag.NewFlagSet(appName, flag.ContinueOnError)
	fs.SetOutput(io.Discard) // prevent default stderr output on error so we can format it

	typeFlag := fs.String("type", "", "Type of movies to fetch (playing, popular, top, upcoming)")
	apiKeyFlag := fs.String("api-key", "", "TMDB API Key or Read Access Token")
	pageFlag := fs.Int("page", 1, "Page number to fetch")
	limitFlag := fs.Int("limit", 0, "Maximum number of results to display")
	jsonFlag := fs.Bool("json", false, "Output raw JSON")
	detailedFlag := fs.Bool("detailed", false, "Show detailed movie information")
	versionFlag := fs.Bool("version", false, "Show application version")
	vFlag := fs.Bool("v", false, "Show application version (shorthand)")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	showVersion := *versionFlag || *vFlag

	// If version is requested, we don't require --type or API key
	if showVersion {
		return Config{ShowVersion: true}, nil
	}

	rawType := strings.TrimSpace(*typeFlag)
	if rawType == "" {
		return Config{}, fmt.Errorf("missing required flag: --type\nUsage: %s --type <playing|popular|top|upcoming>", appName)
	}

	movieType, ok := ParseMovieType(rawType)
	if !ok {
		return Config{}, fmt.Errorf("invalid movie type %q: supported types are 'playing', 'popular', 'top', 'upcoming'", rawType)
	}

	page := *pageFlag
	if page < 1 {
		return Config{}, fmt.Errorf("invalid page number: %d (page must be >= 1)", page)
	}

	limit := *limitFlag
	if limit < 0 {
		return Config{}, fmt.Errorf("invalid limit: %d (limit must be >= 0)", limit)
	}

	// Resolve API key from flag or environment variables
	apiKey := strings.TrimSpace(*apiKeyFlag)
	if apiKey == "" {
		apiKey = os.Getenv("TMDB_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("TMDB_TOKEN")
	}
	if apiKey == "" {
		apiKey = os.Getenv("TMDB_BEARER_TOKEN")
	}

	return Config{
		Type:        rawType,
		MovieType:   movieType,
		APIKey:      strings.TrimSpace(apiKey),
		Page:        page,
		Limit:       limit,
		JSON:        *jsonFlag,
		Detailed:    *detailedFlag,
		ShowVersion: false,
	}, nil
}
