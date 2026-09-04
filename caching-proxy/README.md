# Caching Proxy

A caching reverse proxy CLI that forwards HTTP requests to an origin server and caches responses in Redis. Subsequent identical requests are served from cache without hitting the origin.

Project URL: https://roadmap.sh/projects/caching-server

## Features

- Forwards all HTTP methods to a configurable origin server
- Caches 2xx responses in Redis with a 5-minute TTL
- Sets `X-Cache: HIT` or `X-Cache: MISS` on every response
- `--clear-cache` flag to flush the cache and exit
- Validates CLI flags and environment variables on startup
- Graceful shutdown on `SIGINT`/`SIGTERM`

## Requirements

- Go 1.21+
- Redis

## Setup

```sh
git clone https://github.com/xenptr/go-projects
cd go-projects/caching-proxy
cp .env.example .env
```

Edit `.env` with your Redis connection details:

```env
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_USERNAME=
REDIS_PASSWORD=
```

## Usage

**Start the proxy:**

```sh
go run ./cmd/cli --port <port> --origin <url>
```

```sh
go run ./cmd/cli --port 3000 --origin http://dummyjson.com
```

**Clear the cache:**

```sh
go run ./cmd/cli --clear-cache
```

**Help:**

```sh
go run ./cmd/cli --help
```

## How It Works

```
Client → Proxy (checks Redis)
              ↓ MISS           ↓ HIT
         Origin server     Return cached response
              ↓
         Cache response (2xx only)
              ↓
         Return response
```

1. Request comes in — proxy builds a cache key from the HTTP method and request URI.
2. If the key exists in Redis, the cached response is returned with `X-Cache: HIT`.
3. On a miss, the request is forwarded to the origin. The response is stored in Redis if the status is 2xx, then returned with `X-Cache: MISS`.

## Project Structure

```
caching-proxy/
├── cmd/cli/
│   └── main.go          # Entry point, wires config → redis → proxy
├── internal/
│   ├── cache/
│   │   └── redis.go     # Redis client setup and health check
│   ├── config/
│   │   └── config.go    # CLI flag parsing and env var loading
│   └── proxy/
│       └── proxy.go     # Reverse proxy with cache logic
├── .env.example
└── go.mod
```
