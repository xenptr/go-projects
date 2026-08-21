package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xenptr/go-projects/expense-tracker-api/internal/ratelimit"
)

type mockLimiter struct {
	allowFunc func(ctx context.Context, key string, cfg ratelimit.Config) (ratelimit.Result, error)
}

func (m *mockLimiter) Allow(ctx context.Context, key string, cfg ratelimit.Config) (ratelimit.Result, error) {
	if m.allowFunc != nil {
		return m.allowFunc(ctx, key, cfg)
	}
	return ratelimit.Result{Allowed: true, Limit: cfg.Limit, Remaining: cfg.Limit - 1}, nil
}

func TestRateLimit(t *testing.T) {
	cfg := ratelimit.Config{
		Limit:  10,
		Window: time.Minute,
	}

	tests := []struct {
		name               string
		limiter            ratelimit.Limiter
		expectedStatus     int
		expectNext         bool
		expectedLimitHdr   string
		expectedRemainHdr  string
		expectedRetryAfter string
	}{
		{
			name: "request allowed within limit",
			limiter: &mockLimiter{
				allowFunc: func(ctx context.Context, key string, cfg ratelimit.Config) (ratelimit.Result, error) {
					return ratelimit.Result{
						Allowed:   true,
						Limit:     10,
						Remaining: 9,
					}, nil
				},
			},
			expectedStatus:    http.StatusOK,
			expectNext:        true,
			expectedLimitHdr:  "10",
			expectedRemainHdr: "9",
		},
		{
			name: "request rate limited over limit",
			limiter: &mockLimiter{
				allowFunc: func(ctx context.Context, key string, cfg ratelimit.Config) (ratelimit.Result, error) {
					return ratelimit.Result{
						Allowed:    false,
						Limit:      10,
						Remaining:  0,
						RetryAfter: 5 * time.Second,
					}, nil
				},
			},
			expectedStatus:     http.StatusTooManyRequests,
			expectNext:         false,
			expectedLimitHdr:   "10",
			expectedRemainHdr:  "0",
			expectedRetryAfter: "5",
		},
		{
			name: "limiter error fails open",
			limiter: &mockLimiter{
				allowFunc: func(ctx context.Context, key string, cfg ratelimit.Config) (ratelimit.Result, error) {
					return ratelimit.Result{}, errors.New("redis unavailable")
				},
			},
			expectedStatus: http.StatusOK,
			expectNext:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calledNext := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calledNext = true
				w.WriteHeader(http.StatusOK)
			})

			mw := RateLimit(tt.limiter, cfg, "test")(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/resource", nil)
			req.RemoteAddr = "192.0.2.1:12345"
			rec := httptest.NewRecorder()

			mw.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if calledNext != tt.expectNext {
				t.Fatalf("calledNext = %v, want %v", calledNext, tt.expectNext)
			}

			if tt.expectedLimitHdr != "" && rec.Header().Get("X-RateLimit-Limit") != tt.expectedLimitHdr {
				t.Errorf("X-RateLimit-Limit = %s, want %s", rec.Header().Get("X-RateLimit-Limit"), tt.expectedLimitHdr)
			}

			if tt.expectedRemainHdr != "" && rec.Header().Get("X-RateLimit-Remaining") != tt.expectedRemainHdr {
				t.Errorf("X-RateLimit-Remaining = %s, want %s", rec.Header().Get("X-RateLimit-Remaining"), tt.expectedRemainHdr)
			}

			if tt.expectedRetryAfter != "" && rec.Header().Get("Retry-After") != tt.expectedRetryAfter {
				t.Errorf("Retry-After = %s, want %s", rec.Header().Get("Retry-After"), tt.expectedRetryAfter)
			}
		})
	}
}
