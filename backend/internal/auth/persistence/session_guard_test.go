package persistence

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/auth/domain"
)

type stubTokens struct {
	domain.TokenRepository
	revokedAt time.Time
	err       error
}

func (s stubTokens) SessionsRevokedAt(context.Context, int) (time.Time, error) {
	return s.revokedAt, s.err
}

type stubActive struct {
	active bool
	err    error
	calls  int
}

func (s *stubActive) IsActive(context.Context, int) (bool, error) {
	s.calls++
	return s.active, s.err
}

func TestGuardRejectsTokensIssuedBeforeRevocation(t *testing.T) {
	now := time.Now()
	g := NewSessionGuard(stubTokens{revokedAt: now}, &stubActive{active: true})
	if g.Valid(context.Background(), 1, now.Add(-time.Minute)) {
		t.Fatal("token issued before the revocation must be rejected")
	}
	if !g.Valid(context.Background(), 1, now.Add(time.Second)) {
		t.Fatal("token issued after the revocation must pass")
	}
}

func TestGuardRejectsDeactivatedAccountsAndCaches(t *testing.T) {
	users := &stubActive{active: false}
	g := NewSessionGuard(stubTokens{}, users)
	if g.Valid(context.Background(), 1, time.Now()) || g.Valid(context.Background(), 1, time.Now()) {
		t.Fatal("deactivated account must be rejected")
	}
	if users.calls != 1 {
		t.Fatalf("lookup must be cached, got %d calls", users.calls)
	}
	g.Invalidate(1)
	users.active = true
	if !g.Valid(context.Background(), 1, time.Now()) {
		t.Fatal("after invalidation the fresh state must be read")
	}
}

func TestGuardCacheExpires(t *testing.T) {
	users := &stubActive{active: true}
	g := NewSessionGuard(stubTokens{}, users)
	clock := time.Now()
	g.now = func() time.Time { return clock }
	g.Valid(context.Background(), 1, clock)
	clock = clock.Add(activeCacheTTL + time.Second)
	users.active = false
	if g.Valid(context.Background(), 1, clock) {
		t.Fatal("an expired cache entry must be refreshed")
	}
}

func TestGuardFailsOpenOnBackendErrors(t *testing.T) {
	g := NewSessionGuard(stubTokens{err: errors.New("redis down")}, &stubActive{err: errors.New("db down")})
	if !g.Valid(context.Background(), 1, time.Now()) {
		t.Fatal("backend errors must fail open")
	}
}

// Above activeCacheMax distinct users the cache must evict, not wipe itself: a wipe would send the
// next requests of every still-active user to the database.
func TestGuardCacheEvictsInsteadOfWiping(t *testing.T) {
	users := &stubActive{active: true}
	g := NewSessionGuard(stubTokens{}, users)
	g.now = func() time.Time { return time.Unix(1_000_000, 0) }

	for id := 1; id <= activeCacheMax; id++ {
		g.Valid(context.Background(), id, g.now())
	}
	if len(g.cache) != activeCacheMax {
		t.Fatalf("cache size %d, want %d", len(g.cache), activeCacheMax)
	}
	g.Valid(context.Background(), activeCacheMax+1, g.now())
	if len(g.cache) != activeCacheMax {
		t.Fatalf("cache size after overflow %d, want %d (one entry evicted)", len(g.cache), activeCacheMax)
	}
	if _, ok := g.cache[activeCacheMax+1]; !ok {
		t.Fatal("the new entry must be cached")
	}
}

func TestGuardCacheDropsExpiredEntriesWhenFull(t *testing.T) {
	users := &stubActive{active: true}
	g := NewSessionGuard(stubTokens{}, users)
	now := time.Unix(1_000_000, 0)
	g.now = func() time.Time { return now }
	for id := 1; id <= activeCacheMax; id++ {
		g.Valid(context.Background(), id, now)
	}
	now = now.Add(2 * activeCacheTTL) // everything expired
	g.Valid(context.Background(), activeCacheMax+1, now)
	if len(g.cache) != 1 {
		t.Fatalf("cache size %d, want 1 after dropping expired entries", len(g.cache))
	}
}
