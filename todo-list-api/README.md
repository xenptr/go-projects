# Todo List API

A RESTful API for managing personal to-do lists with user authentication, JWT-based auth with refresh token rotation, rate limiting, and PostgreSQL persistence.

Built as part of the [roadmap.sh backend projects](https://roadmap.sh/projects/todo-list-api).

## Features

- User registration and login
- JWT access tokens (15 min) + refresh tokens (7 days) with rotation
- Full CRUD for to-do items
- Ownership checks — users can only modify their own todos
- Pagination and search on the list endpoint
- Per-IP rate limiting via Redis
- PostgreSQL for storage, Redis for refresh tokens and rate limiting

## Tech Stack

- Go (standard library `net/http`)
- PostgreSQL + pgx
- Redis
- golang-jwt/jwt

## Project Structure

```
cmd/api/          → main entry point
internal/
  auth/           → JWT generation/parsing, password hashing, refresh token store
  config/         → env-based config loading
  db/             → pgx connection pool setup
  dto/            → request/response structs
  handlers/       → HTTP handlers
  middleware/     → auth and rate limit middleware
  models/         → domain models (User, Todo)
  ratelimit/      → Redis fixed-window rate limiter
  redis/          → Redis client setup
  repository/     → PostgreSQL queries
  routes/         → route registration
  validation/     → input validation
sql/migrations/   → up/down SQL migration files
```

## Getting Started

### Prerequisites

- Go 1.22+
- PostgreSQL
- Redis

### Setup

1. Clone the repo and navigate into the project:

```bash
cd todo-list-api
```

2. Copy the example env file and fill in your values:

```bash
cp .env.example .env
```

```env
APP_PORT=8080

DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=todo_api

JWT_SECRET=your_super_secret_key

REDIS_HOST=127.0.0.1
REDIS_PORT=6379
```

3. Run the database migrations. The SQL files are in `sql/migrations/` — run the `.up.sql` files in order:

```bash
psql -U your_db_user -d todo_api -f sql/migrations/000001_create_users.up.sql
psql -U your_db_user -d todo_api -f sql/migrations/000002_create_todos.up.sql
```

4. Start the server:

```bash
go run ./cmd/api
```

The server will start on the port defined in `APP_PORT` (default `8080`).

## API Endpoints

### Auth

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/register` | Register a new user |
| POST | `/login` | Login and get tokens |
| POST | `/auth/refresh` | Refresh access token |

### Todos (require `Authorization: Bearer <token>`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/todos` | List todos (paginated) |
| POST | `/todos` | Create a todo |
| PUT | `/todos/{id}` | Update a todo |
| DELETE | `/todos/{id}` | Delete a todo |

### Register

```
POST /register
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123"
}
```

Response:

```json
{
  "token": "eyJhbGci...",
  "refresh_token": "eyJhbGci..."
}
```

### Login

```
POST /login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "password123"
}
```

### Create Todo

```
POST /todos
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Buy groceries",
  "description": "Milk, eggs, bread"
}
```

### List Todos

```
GET /todos?page=1&limit=10&search=groceries
Authorization: Bearer <token>
```

Response:

```json
{
  "data": [
    {
      "id": 1,
      "title": "Buy groceries",
      "description": "Milk, eggs, bread",
      "copmleted": false,
      "created_at": "2026-08-20T10:00:00Z"
    }
  ],
  "page": 1,
  "limit": 10,
  "total": 1
}
```

### Update Todo

```
PUT /todos/1
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Buy groceries",
  "description": "Milk, eggs, bread, and cheese",
  "completed": true
}
```

### Delete Todo

```
DELETE /todos/1
Authorization: Bearer <token>
```

Returns `204 No Content` on success.

### Refresh Token

```
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGci..."
}
```

## Error Responses

All errors return JSON with a `message` field:

```json
{ "message": "Unauthorized" }
```

| Status | Meaning |
|--------|---------|
| 400 | Bad request / validation failed |
| 401 | Missing or invalid token |
| 403 | Authenticated but not the owner |
| 404 | Resource not found |
| 409 | Email already in use |
| 429 | Rate limit exceeded |
| 500 | Internal server error |

## Rate Limiting

- Auth endpoints (`/register`, `/login`, `/auth/refresh`): 5 requests per minute per IP
- Todo endpoints: 100 requests per minute per IP

Uses a Redis fixed-window algorithm. Responses include `X-RateLimit-Limit` and `X-RateLimit-Remaining` headers.
