package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter interface {
	Allow(ctx context.Context, key string, cfg Config) (Result, error)
}

type RedisLimiter struct {
	client *redis.Client
}

func New(client *redis.Client) *RedisLimiter {
	return &RedisLimiter{
		client: client,
	}
}

type Config struct {
	Limit  int
	Window time.Duration
}

// Result carries enough info for the middleware to set standard
type Result struct {
	Allowed    bool
	Limit      int
	Remaining  int
	RetryAfter time.Duration
}

// fixedWindowScript increments a counter and sets its expiry only on
// the first hit in the window, all atomically server-side. Returns the
// count after incrementing and the TTL remaining (ms).
var fixedWindowScript = redis.NewScript(`
local key = KEYS[1]
local window_ms = tonumber(ARGV[1])
 
local count = redis.call("INCR", key)

if count == 1 then
    redis.call("PEXPIRE", key, window_ms)
end

local ttl = redis.call("PTTL", key)
 
return {count, ttl}
`)

func (l *RedisLimiter) Allow(ctx context.Context, key string, cfg Config) (Result, error) {
	if cfg.Limit <= 0 {
		return Result{}, errors.New("ratelimit: limit must be greater than zero")
	}

	if cfg.Window <= 0 {
		return Result{}, errors.New("ratelimit: window must be greater than zero")
	}

	redisKey := "ratelimit:" + key

	res, err := fixedWindowScript.Run(
		ctx,
		l.client,
		[]string{redisKey},
		cfg.Window.Milliseconds(),
	).Result()
	if err != nil {
		return Result{}, fmt.Errorf("rate limit script: %w", err)
	}

	vals, ok := res.([]interface{})
	if !ok || len(vals) != 2 {
		return Result{}, errors.New("ratelimit: unexpected script result")
	}

	count, ok := vals[0].(int64)
	if !ok {
		return Result{}, errors.New("ratelimit: invalid count")
	}

	ttlMs, ok := vals[1].(int64)
	if !ok {
		return Result{}, errors.New("ratelimit: invalid ttl")
	}

	remaining := max(0, cfg.Limit-int(count))

	return Result{
		Allowed:    count <= int64(cfg.Limit),
		Limit:      cfg.Limit,
		Remaining:  remaining,
		RetryAfter: time.Duration(ttlMs) * time.Millisecond,
	}, nil
}
