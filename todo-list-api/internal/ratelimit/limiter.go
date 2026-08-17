package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter struct {
	client *redis.Client
}

func New(client *redis.Client) *Limiter {
	return &Limiter{
		client: client,
	}
}

type Config struct {
	Limit  int
	Window time.Duration
}

func (l *Limiter) Allow(ctx context.Context, key string, cfg Config) (bool, error) {
	count, err := l.client.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("increment rate limit: %w", err)
	}

	// First request in this window.
	if count == 1 {
		if err := l.client.Expire(ctx, key, cfg.Window).Err(); err != nil {
			return false, fmt.Errorf("set rate limit expiration: %w", err)
		}
	}

	return count <= int64(cfg.Limit), nil
}
