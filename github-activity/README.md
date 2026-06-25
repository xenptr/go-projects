# GitHub Activity CLI

A command-line tool that fetches and displays the recent public activity of any GitHub user using the GitHub Events API.

## Project URL

https://roadmap.sh/projects/github-user-activity

## Features

- Fetch recent public activity for any GitHub user
- Display human-readable summaries of GitHub events including:
  - Push events
  - Pull request opens, closes, merges, reviews, and comments
  - Issue opens, closes, assignments, and labels
  - Repository creates, forks, and deletes
  - Stars (watch events)
  - Releases
  - Member (collaborator) changes
  - Commit comments

## Installation

Clone the repository:

```bash
git clone <repository-url>
cd github-activity
```

Run directly:

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
| `--json` | Output raw JSON |
| `--types` | List all supported event types |

Examples:

```bash
./github-activity torvalds
./github-activity torvalds --limit 10
./github-activity torvalds --filter PushEvent
./github-activity torvalds --limit 5 --filter PushEvent
./github-activity torvalds --json
```

If the user has no recent public activity:

```
No recent public activity found for <username>
```

## Project Structure

```text
.
├── main.go         # Entry point, flag parsing, output loop
├── github.go       # GitHub API client and event fetching
├── formatter.go    # Event formatting and human-readable output
├── config.go       # CLI flag definitions and Config struct
├── types.go        # Event, Payload, and related type definitions
├── eventtypes.go   # Supported event type list
├── go.mod
└── README.md
```

## API

Uses the public [GitHub Events API](https://docs.github.com/en/rest/activity/events):

```
GET https://api.github.com/users/{username}/events?per_page={limit}
```

No authentication is required, but unauthenticated requests are subject to GitHub's rate limits.
