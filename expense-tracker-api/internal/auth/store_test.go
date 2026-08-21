package auth

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// inMemoryRefreshStore is a thread-safe in-memory implementation of
// RefreshTokenStore used exclusively in tests.
type inMemoryRefreshStore struct {
	mu     sync.Mutex
	tokens map[string]struct{}
}

func newInMemoryRefreshStore() *inMemoryRefreshStore {
	return &inMemoryRefreshStore{tokens: make(map[string]struct{})}
}

func storeKey(userID int64, token string) string {
	return fmt.Sprintf("%d:%s", userID, token)
}

func (s *inMemoryRefreshStore) Save(_ context.Context, userID int64, token string, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[storeKey(userID, token)] = struct{}{}
	return nil
}

func (s *inMemoryRefreshStore) Exists(_ context.Context, userID int64, token string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.tokens[storeKey(userID, token)]
	return ok, nil
}

func (s *inMemoryRefreshStore) Revoke(_ context.Context, userID int64, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, storeKey(userID, token))
	return nil
}

func TestRefreshTokenStore_SaveAndExists(t *testing.T) {
	store := newInMemoryRefreshStore()
	ctx := context.Background()

	ok, err := store.Exists(ctx, 1, "token-abc")
	if err != nil {
		t.Fatalf("Exists() returned error: %v", err)
	}
	if ok {
		t.Fatal("expected token to not exist before Save")
	}

	if err = store.Save(ctx, 1, "token-abc", time.Minute); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	ok, err = store.Exists(ctx, 1, "token-abc")
	if err != nil {
		t.Fatalf("Exists() returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected token to exist after Save")
	}
}

func TestRefreshTokenStore_Revoke(t *testing.T) {
	store := newInMemoryRefreshStore()
	ctx := context.Background()

	_ = store.Save(ctx, 2, "token-xyz", time.Minute)

	if err := store.Revoke(ctx, 2, "token-xyz"); err != nil {
		t.Fatalf("Revoke() returned error: %v", err)
	}

	ok, err := store.Exists(ctx, 2, "token-xyz")
	if err != nil {
		t.Fatalf("Exists() returned error: %v", err)
	}
	if ok {
		t.Fatal("expected token to not exist after Revoke")
	}
}

func TestRefreshTokenStore_ScopedToUser(t *testing.T) {
	store := newInMemoryRefreshStore()
	ctx := context.Background()

	_ = store.Save(ctx, 10, "shared-token", time.Minute)

	// Same token string, different user — must not exist.
	ok, err := store.Exists(ctx, 99, "shared-token")
	if err != nil {
		t.Fatalf("Exists() returned error: %v", err)
	}
	if ok {
		t.Fatal("token should be scoped to its owner user, not visible to other users")
	}
}
