# Expense Tracker API

A RESTful API for tracking personal expenses, built with Go. This is one of the backend projects from [roadmap.sh](https://roadmap.sh/projects/expense-tracker-api).

## What it does

Users can sign up, log in, and manage their own expenses. Everything is scoped per user — you only ever see and touch your own records. Expenses have a title, amount, date, category, and an optional description. You can also filter your expense list by time periods like the past week, past month, or a custom date range.

Authentication uses short-lived JWT access tokens (15 minutes) paired with long-lived refresh tokens (7 days) stored in Redis. Refresh tokens are rotated on each use, so reusing an old one gets rejected immediately.

## Features

- User registration and login
- JWT-based auth with access + refresh token rotation
- Create, read, update, and delete expenses
- Partial updates — only send the fields you want to change
- Filter expenses by: past week, past month, last 3 months, or a custom date range
- Expense categories: Groceries, Leisure, Electronics, Utilities, Clothing, Health, Others
- Rate limiting: auth routes (5 req/min), expense routes (100 req/min)
- Graceful shutdown

## Stack

- Go (standard library `net/http` with 1.22+ routing patterns)
- PostgreSQL with `pgx/v5`
- Redis for refresh token storage and rate limiting
- `golang-jwt/jwt` for token generation and validation
- `shopspring/decimal` for precise money handling

## Getting started

### Prerequisites

- Go 1.22+
- PostgreSQL
- Redis

### Setup

1. Clone the repo and navigate into the project directory.

2. Copy the example env file and fill in your values:

```
cp .env.example .env
```

3. Create the database and run the schema:

```
psql -U your_user -d your_db -f sql/schema.sql
```

4. Run the server:

```
go run ./cmd/api
```

The server starts on port 8080 by default. Change it with `APP_PORT` in `.env`.

## Environment variables

| Variable | Description | Default |
|---|---|---|
| `APP_PORT` | Port to listen on | `8080` |
| `DB_HOST` | PostgreSQL host | `127.0.0.1` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | — |
| `DB_PASSWORD` | Database password | — |
| `DB_NAME` | Database name | — |
| `JWT_SECRET` | Secret key for signing JWTs | — |
| `REDIS_HOST` | Redis host | — |
| `REDIS_PORT` | Redis port | — |
| `REDIS_USERNAME` | Redis username (optional) | — |
| `REDIS_PASSWORD` | Redis password (optional) | — |

## API

### Auth

#### Register
```
POST /register
Content-Type: application/json

{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "secret123"
}
```

Returns a `201` with an access token and refresh token.

#### Login
```
POST /login
Content-Type: application/json

{
  "email": "jane@example.com",
  "password": "secret123"
}
```

Returns a `200` with an access token and refresh token.

#### Refresh token
```
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "<your refresh token>"
}
```

Returns a new token pair and invalidates the old refresh token.

---

All expense routes require the `Authorization: Bearer <access_token>` header.

### Expenses

#### List expenses
```
GET /expenses
GET /expenses?filter=week
GET /expenses?filter=month
GET /expenses?filter=3months
GET /expenses?filter=custom&start_date=2024-01-01&end_date=2024-03-31
```

Returns an array of your expenses, newest first.

#### Create an expense
```
POST /expenses
Content-Type: application/json

{
  "title": "Groceries run",
  "amount": "54.30",
  "category": "Groceries",
  "date": "2024-07-15",
  "description": "weekly shop"
}
```

`category` defaults to `Others` and `date` defaults to today if omitted. Returns `201` with the created expense.

#### Get an expense
```
GET /expenses/{id}
```

#### Update an expense
```
PUT /expenses/{id}
Content-Type: application/json

{
  "amount": "60.00"
}
```

All fields are optional — only the ones you include get updated.

#### Delete an expense
```
DELETE /expenses/{id}
```

Returns `204 No Content` on success.

---

### Response format

Successful responses return the resource as JSON. Errors look like:

```json
{
  "error": "expense not found"
}
```

Validation errors return a `422` with a map of field-level messages:

```json
{
  "errors": {
    "amount": "must be a positive number",
    "category": "invalid category"
  }
}
```

## Running tests

```
go test ./...
```
