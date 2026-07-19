# Weather API

A simple HTTP API that returns current weather data for a given city. Built with Go and backed by the Visual Crossing weather service. Responses are cached in Redis to avoid hammering the upstream API on repeated requests.

## Project URL

https://roadmap.sh/projects/weather-api-wrapper-service

## How it works

Hit `/{city}` and you get back the weather data as JSON. On the first request for a city the server fetches from Visual Crossing and stores the result in Redis for 12 hours. Subsequent requests for the same city are served straight from cache. There's also a basic rate limiter — each IP is capped at 100 requests per minute.

## Requirements

- A [Visual Crossing](https://www.visualcrossing.com/) API key (free tier works fine)
- A running Redis instance

## Setup

Set the required environment variables:

```bash
export VISUAL_CROSSING_API_KEY=your_key_here
export REDIS_HOST=127.0.0.1
export REDIS_PORT=6379

# optional, only if your Redis needs auth
export REDIS_USERNAME=
export REDIS_PASSWORD=
```

Then run:

```bash
go run .
```

The server starts on `localhost:8080`.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Health check, returns a welcome message |
| GET | `/{city}` | Returns weather data for the given city |

Example:

```bash
curl http://localhost:8080/London
```

## Project Structure

```text
.
├── main.go     # Server, handlers, cache and rate limiter logic
├── go.mod
├── go.sum
└── README.md
```
