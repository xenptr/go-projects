# GitHub Activity CLI

A small CLI tool to check what a GitHub user has been up to lately. It hits the GitHub Events API and prints a readable summary of their recent public activity.

## Project URL

https://roadmap.sh/projects/github-user-activity

## Features

- View a user's recent public GitHub activity
- Display events in a readable format
- Filter events by type
- Profile view with basic user info (`--profile`)
- Output raw API responses as JSON (`--json`)
- Cache responses for five minutes to reduce API requests (`--cache`)

## Installation

Clone the repository:

```bash
git clone <repository-url>
cd github-activity
```

Run without building:

```bash
go run . <username>
```

Or build an executable:

```bash
go build -o github-activity
```

Run the executable:

```bash
./github-activity <username>
```

## Usage

```
github-activity <username> [options]
```

| Flag | Description |
|------|-------------|
| `--filter <type>` | Filter by event type (e.g. `PushEvent`) |
| `--limit <n>` | Limit the number of events fetched (0 = no limit, max 100) |
| `--json` | Output as JSON instead of formatted text |
| `--profile` | Show the user's GitHub profile info |
| `--cache` | Cache API responses for 5 minutes |
| `--types` | List all supported event types |

Examples:

```bash
./github-activity torvalds
./github-activity torvalds --limit 10
./github-activity torvalds --filter PushEvent
./github-activity torvalds --profile
./github-activity torvalds --profile --json
./github-activity torvalds --limit 5 --cache
./github-activity torvalds --json
```

## Project Structure

```text
.
├── main.go         # Entry point and output logic
├── github.go       # GitHub API client
├── cache.go        # File-based response caching
├── formatter.go    # Formats events for display
├── config.go       # CLI flags and Config struct
├── types.go        # API response structs
├── eventtypes.go   # Supported GitHub event types
├── go.mod
└── README.md
```

## API

Uses the public [GitHub Events API](https://docs.github.com/en/rest/activity/events):

```
GET https://api.github.com/users/{username}/events?per_page={limit}
GET https://api.github.com/users/{username}
```

No authentication needed, but unauthenticated requests are rate-limited by GitHub (60 requests/hour). Use `--cache` to stay well within that.
