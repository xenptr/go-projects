package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RefreshTokenStore manages the persistence and revocation of refresh tokens.
// Storing tokens in Redis allows instant revocation (logout, rotation) without
// waiting for the JWT expiry time.
type RefreshTokenStore interface {
	// Save persists a refresh token for the given user with a TTL.
	Save(ctx context.Context, userID int64, token string, ttl time.Duration) error

	// Exists reports whether the given refresh token is still valid (not revoked
	// and not expired).
	Exists(ctx context.Context, userID int64, token string) (bool, error)

	// Revoke deletes a specific refresh token, effectively logging out that session.
	Revoke(ctx context.Context, userID int64, token string) error
}

// redisRefreshStore is the production Redis-backed implementation of RefreshTokenStore.
type redisRefreshStore struct {
	client *redis.Client
}

// NewRedisRefreshStore returns a RefreshTokenStore backed by Redis.
func NewRedisRefreshStore(client *redis.Client) RefreshTokenStore {
	return &redisRefreshStore{client: client}
}

// key builds a namespaced Redis key for a user+token pair.
func (s *redisRefreshStore) key(userID int64, token string) string {
	return fmt.Sprintf("refresh_token:%d:%s", userID, token)
}

func (s *redisRefreshStore) Save(ctx context.Context, userID int64, token string, ttl time.Duration) error {
	return s.client.Set(ctx, s.key(userID, token), "1", ttl).Err()
}

func (s *redisRefreshStore) Exists(ctx context.Context, userID int64, token string) (bool, error) {
	n, err := s.client.Exists(ctx, s.key(userID, token)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *redisRefreshStore) Revoke(ctx context.Context, userID int64, token string) error {
	return s.client.Del(ctx, s.key(userID, token)).Err()
}
