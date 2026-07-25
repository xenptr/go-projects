# Blogging Platform API

A RESTful JSON API for managing blog posts, backed by PostgreSQL. Supports creating, reading, updating, and deleting posts, plus a full-text search across title, content, category, and tags.

The project has two store implementations — one using `database/sql` with the pgx stdlib driver, and one using the pgx native API with a connection pool (`pgxpool`). The server currently runs on the native pgx implementation.

## Project URL

https://roadmap.sh/projects/blogging-platform-api

## Requirements

- Go 1.22 or later (uses `net/http` path value routing)
- A running PostgreSQL instance

## Setup

Create the database table by running the schema:

```bash
psql -U your_user -d your_db -f sql/schema.sql
```

Set the environment variables:

```bash
export DBHOST=127.0.0.1
export DBPORT=5432
export DBUSER=your_user
export DBPASS=your_password
export DBNAME=your_db

# optional, defaults to 8080
export APP_PORT=8080
```

Then start the server:

```bash
go run ./cmd/api
```

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/posts` | List all posts |
| GET | `/posts?term=go` | Search posts by keyword |
| GET | `/posts/{id}` | Get a single post |
| POST | `/posts` | Create a new post |
| PUT | `/posts/{id}` | Update an existing post |
| DELETE | `/posts/{id}` | Delete a post |

### Post fields

| Field | Type | Required |
|-------|------|----------|
| title | string | yes |
| content | string | yes |
| category | string | no |
| tags | array of strings | no |

### Examples

Create a post:

```bash
curl -X POST http://localhost:8080/posts \
  -H "Content-Type: application/json" \
  -d '{"title":"Hello World","content":"My first post","category":"general","tags":["go","api"]}'
```

Search posts:

```bash
curl "http://localhost:8080/posts?term=go"
```

Get a post by id:

```bash
curl http://localhost:8080/posts/1
```

## Project Structure

```text
.
├── cmd/
│   └── api/
│       └── main.go                 # Entry point
├── internal/
│   ├── config/
│   │   └── config.go               # Loads config from env vars
│   ├── database/
│   │   ├── postgres.go             # database/sql connection (pgx stdlib driver)
│   │   └── pgx.go                  # pgxpool connection (pgx native)
│   ├── handlers/
│   │   ├── handler.go              # Handler struct, depends on PostStore interface
│   │   ├── posts.go                # Post CRUD handlers
│   │   ├── root.go                 # Health check handler
│   │   └── routes.go               # Route registration
│   ├── models/
│   │   └── post.go                 # Post struct
│   └── store/
│       ├── store.go                # PostStore interface
│       ├── errors.go               # Sentinel errors shared across implementations
│       ├── pgx/
│       │   └── post_store.go       # pgx native implementation (currently in use)
│       └── sql/
│           └── post_store.go       # database/sql implementation
├── sql/
│   └── schema.sql                  # Table definition
├── go.mod
├── go.sum
└── README.md
```
