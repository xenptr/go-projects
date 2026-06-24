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

Pass a GitHub username as the only argument:

```bash
./github-activity torvalds
```

Example output:

```
- Pushed changes to torvalds/linux
- Opened issue #1234 in torvalds/linux
- Merged PR #5678 in torvalds/linux
- Starred someuser/somerepo
```

If the user has no recent public activity:

```
No recent public activity found for <username>
```

## Project Structure

```text
.
├── main.go       # Entry point, argument parsing, output loop
├── github.go     # GitHub API client and event fetching
├── display.go    # Event formatting and human-readable output
├── types.go      # Event, Payload, and related type definitions
├── go.mod
└── README.md
```

## API

Uses the public [GitHub Events API](https://docs.github.com/en/rest/activity/events):

```
GET https://api.github.com/users/{username}/events
```

No authentication is required, but unauthenticated requests are subject to GitHub's rate limits.
