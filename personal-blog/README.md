# Personal Blog

A personal blog built with Go's standard library. Supports writing and publishing articles through a clean web UI, with a protected admin section for managing content.

## Project URL

https://roadmap.sh/projects/personal-blog

## Features

- Public blog with home page listing all articles and individual article pages
- Password-protected admin dashboard to create, edit, and delete articles
- Articles persisted as JSON files — no database required
- HTML templates with a shared layout using Go's `text/template`
- Static file serving for CSS

## Installation

Clone the repository:

```bash
git clone <repository-url>
cd personal-blog
```

Run the server:

```bash
go run .
```

Then open your browser at [http://localhost:8080](http://localhost:8080).

Or build an executable first:

```bash
go build -o personal-blog
./personal-blog
```

## Usage

### Guest section

| URL | Description |
|-----|-------------|
| `/` | Home page — lists all articles sorted by date |
| `/article/{id}` | View a single article |

### Admin section

| URL | Description |
|-----|-------------|
| `/admin` | Dashboard — manage all articles |
| `/admin/new` | Create a new article |
| `/admin/edit/{id}` | Edit an existing article |
| `/admin/delete/{id}` | Delete an article |

The admin section is protected by HTTP Basic Auth. Default credentials: `Admin` / `admin`.

## Project Structure

```text
.
├── main.go         # Entry point — starts server
├── models.go       # Data types (Article, Blog)
├── handlers.go     # HTTP handler functions
├── routes.go       # Route registration
├── middleware.go   # Auth middleware
├── storage.go      # File I/O — load, save, delete articles
├── templates.go    # Template loading and rendering helpers
├── utils.go        # Shared helpers (validation, date formatting)
├── data/
│   └── articles/   # Article JSON files
├── static/
│   └── css/        # Stylesheets
├── templates/      # HTML templates
├── go.mod
└── README.md
```
