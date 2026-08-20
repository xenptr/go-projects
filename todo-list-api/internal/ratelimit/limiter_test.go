package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func newTestRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: ":6379",
	})
}

func newTestLimiter(t *testing.T, key string) (*RedisLimiter, context.Context) {
	t.Helper()

	client := newTestRedis()
	ctx := context.Background()

	t.Cleanup(func() {
		if err := client.Del(ctx, "ratelimit:"+key).Err(); err != nil {
			t.Errorf("failed to clean up redis key: %v", err)
		}

		if err := client.Close(); err != nil {
			t.Errorf("failed to close redis client: %v", err)
		}
	})

	return New(client), ctx
}

func TestRedisLimiter_Allow(t *testing.T) {
	key := "test-user"

	limiter, ctx := newTestLimiter(t, key)

	cfg := Config{
		Limit:  3,
		Window: time.Second,
	}

	t.Run("allows requests within limit", func(t *testing.T) {
		for i := 1; i <= 3; i++ {
			result, err := limiter.Allow(ctx, key, cfg)
			if err != nil {
				t.Fatalf("Allow() returned error: %v", err)
			}

			if !result.Allowed {
				t.Fatalf("request %d should be allowed", i)
			}

			expectedRemaining := 3 - i

			if result.Remaining != expectedRemaining {
				t.Errorf(
					"Remaining = %d, want %d",
					result.Remaining,
					expectedRemaining,
				)
			}

			if result.Limit != 3 {
				t.Errorf("Limit = %d, want 3", result.Limit)
			}
		}
	})

	t.Run("rejects request over limit", func(t *testing.T) {
		result, err := limiter.Allow(ctx, key, cfg)
		if err != nil {
			t.Fatalf("Allow() returned error: %v", err)
		}

		if result.Allowed {
			t.Error("request over limit should not be allowed")
		}

		if result.Remaining != 0 {
			t.Errorf(
				"Remaining = %d, want 0",
				result.Remaining,
			)
		}

		if result.RetryAfter <= 0 {
			t.Error("RetryAfter should be greater than zero")
		}
	})
}

func TestRedisLimiter_Allow_InvalidConfig(t *testing.T) {
	key := "config-test"

	limiter, ctx := newTestLimiter(t, key)

	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "zero limit",
			cfg: Config{
				Limit:  0,
				Window: time.Second,
			},
		},
		{
			name: "negative limit",
			cfg: Config{
				Limit:  -1,
				Window: time.Second,
			},
		},
		{
			name: "zero window",
			cfg: Config{
				Limit:  5,
				Window: 0,
			},
		},
		{
			name: "negative window",
			cfg: Config{
				Limit:  5,
				Window: -time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := limiter.Allow(ctx, key, tt.cfg)

			if err == nil {
				t.Fatal("Allow() should return an error")
			}
		})
	}
}

func TestRedisLimiter_Allow_Limit(t *testing.T) {
	key := "limit-test"

	limiter, ctx := newTestLimiter(t, key)

	cfg := Config{
		Limit:  3,
		Window: time.Second,
	}

	for i := 1; i <= 4; i++ {
		result, err := limiter.Allow(ctx, key, cfg)
		if err != nil {
			t.Fatalf("Allow() returned error: %v", err)
		}

		wantAllowed := i <= 3

		if result.Allowed != wantAllowed {
			t.Errorf(
				"request %d: Allowed = %v, want %v",
				i,
				result.Allowed,
				wantAllowed,
			)
		}

		expectedRemaining := max(0, 3-i)

		if result.Remaining != expectedRemaining {
			t.Errorf(
				"request %d: Remaining = %d, want %d",
				i,
				result.Remaining,
				expectedRemaining,
			)
		}
	}
}

func TestRedisLimiter_Allow_WindowReset(t *testing.T) {
	key := "reset-test"

	limiter, ctx := newTestLimiter(t, key)

	cfg := Config{
		Limit:  1,
		Window: 100 * time.Millisecond,
	}

	result, err := limiter.Allow(ctx, key, cfg)
	if err != nil {
		t.Fatalf("first Allow() returned error: %v", err)
	}

	if !result.Allowed {
		t.Fatal("first request should be allowed")
	}

	result, err = limiter.Allow(ctx, key, cfg)
	if err != nil {
		t.Fatalf("second Allow() returned error: %v", err)
	}

	if result.Allowed {
		t.Fatal("second request should be rejected")
	}

	time.Sleep(150 * time.Millisecond)

	result, err = limiter.Allow(ctx, key, cfg)
	if err != nil {
		t.Fatalf("third Allow() returned error: %v", err)
	}

	if !result.Allowed {
		t.Fatal("request after window reset should be allowed")
	}

	if result.Remaining != 0 {
		t.Errorf(
			"Remaining = %d, want 0",
			result.Remaining,
		)
	}
}

func TestRedisLimiter_Allow_RetryAfter(t *testing.T) {
	key := "retry-test"

	limiter, ctx := newTestLimiter(t, key)

	cfg := Config{
		Limit:  1,
		Window: time.Second,
	}

	_, err := limiter.Allow(ctx, key, cfg)
	if err != nil {
		t.Fatalf("first Allow() returned error: %v", err)
	}

	result, err := limiter.Allow(ctx, key, cfg)
	if err != nil {
		t.Fatalf("second Allow() returned error: %v", err)
	}

	if result.Allowed {
		t.Fatal("second request should be rejected")
	}

	if result.RetryAfter <= 0 {
		t.Error("RetryAfter should be greater than zero")
	}

	if result.RetryAfter > time.Second {
		t.Errorf(
			"RetryAfter = %v, should be <= 1s",
			result.RetryAfter,
		)
	}
}
