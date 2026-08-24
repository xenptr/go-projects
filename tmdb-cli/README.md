# TMDB CLI Tool

A command-line interface (CLI) application to fetch and display movie lists from The Movie Database (TMDB) API right in your terminal.

## Project URL

https://roadmap.sh/projects/tmdb-cli

## Features

- Fetch and view movies by category:
  - **Now Playing** (`playing`): Movies currently in theatres
  - **Popular** (`popular`): Current trending and popular movies
  - **Top Rated** (`top`): All-time top rated movies
  - **Upcoming** (`upcoming`): Movies coming soon to theatres
- Clean, aesthetic terminal display with ratings, votes, release dates, genres, and word-wrapped overviews
- Support for detailed mode (`--detailed`) showing original title, language, age rating, and TMDB IDs
- Pagination support (`--page <n>`) and display limits (`--limit <n>`)
- Raw JSON output mode (`--json`) for scripting and automation
- Secure API key configuration via environment variable or CLI flag
- Robust error handling for rate limits, network timeouts, invalid keys, and API errors
- Built with standard library Go — zero third-party dependencies

## Prerequisites & API Key Setup

To use this tool, you will need a free TMDB API key:

1. Create a free account at [themoviedb.org](https://www.themoviedb.org/signup).
2. Go to **Settings > API** at [themoviedb.org/settings/api](https://www.themoviedb.org/settings/api) and request an API key.
3. Export your API key in your terminal session:

```bash
export TMDB_API_KEY="your_api_key_here"
```

*(You can also pass it directly using the `--api-key` flag).*

## Installation & Build

Clone the repository:

```bash
git clone <repository-url>
cd tmdb-cli
```

Run directly without compiling:

```bash
go run . --type "playing"
```

Or build an executable binary:

```bash
go build -o tmdb-app .
```

## Usage

```text
tmdb-app --type <type> [options]
```

### Movie Types

| Type | Description |
|------|-------------|
| `playing` | Now playing movies in theatres |
| `popular` | Popular movies |
| `top` | Top rated movies of all time |
| `upcoming` | Upcoming movies in theatres |

### Options

| Flag | Description |
|------|-------------|
| `--type <string>` | Movie list type (`playing`, `popular`, `top`, `upcoming`) *(required)* |
| `--api-key <string>` | TMDB API Key or Read Access Token (falls back to `TMDB_API_KEY` env var) |
| `--page <int>` | Page number to fetch (default: `1`) |
| `--limit <int>` | Maximum number of results to display (default: all on page) |
| `--detailed` | Show full overview, genres, original titles, and TMDB metadata |
| `--json` | Output raw JSON data |
| `--version`, `-v` | Show application version |
| `--help`, `-h` | Show help and usage instructions |

## Examples

### View Now Playing Movies
```bash
./tmdb-app --type "playing"
```

### View Popular Movies
```bash
./tmdb-app --type "popular"
```

### View Top Rated Movies (with limit)
```bash
./tmdb-app --type "top" --limit 5
```

### View Upcoming Movies (page 2 with detailed info)
```bash
./tmdb-app --type "upcoming" --page 2 --detailed
```

### Output as JSON
```bash
./tmdb-app --type "playing" --limit 3 --json
```

## Sample Output

```text
──────────────────────────────────────────────────────────────────────────────
🎬 Popular Movies  (Page 1 of 500 · 10,000 total movies)
──────────────────────────────────────────────────────────────────────────────

 1. Dune: Part Two (2024)
    ⭐ Rating: 8.3/10 (5,420 votes)  |  🔥 Popularity: 345.8
    🎭 Genres: Science Fiction, Adventure
    📅 Release: 2024-03-01
    📝 Overview: Follow the mythic journey of Paul Atreides as he unites with
       Chani and the Fremen while seeking revenge against the conspirators who
       destroyed his family.

 2. Inside Out 2 (2024)
    ⭐ Rating: 7.7/10 (3,890 votes)  |  🔥 Popularity: 290.4
    🎭 Genres: Animation, Family, Comedy
    📅 Release: 2024-06-14
    📝 Overview: Teenager Riley's mind headquarters is undergoing a sudden
       demolition to make room for unexpected new Emotions!

──────────────────────────────────────────────────────────────────────────────
Showing 20 movies (Page 1 of 500)
```

## Running Tests

Run the unit test suite with race detection:

```bash
go test -v -race ./...
```

## Project Structure

```text
.
├── main.go           # CLI entry point, argument parsing, and command execution
├── config.go         # Flag definitions, usage formatting, and configuration
├── config_test.go    # Tests for flag parsing, defaults, and validation
├── tmdb.go           # TMDB HTTP API client and error handling
├── tmdb_test.go      # Tests for API client using httptest mock server
├── formatter.go      # Terminal output formatting and text wrapping
├── formatter_test.go # Tests for formatting output
├── types.go          # Data structs, movie types, and genre helpers
├── types_test.go     # Tests for type parsing and genre mapping
├── go.mod            # Go module definition
└── README.md         # Documentation
```
