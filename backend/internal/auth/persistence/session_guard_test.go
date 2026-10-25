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

func TestGuardRejectsTokensIssuedInTheSecondOfTheRevocation(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	g := NewSessionGuard(stubTokens{revokedAt: now}, &stubActive{active: true})
	if g.Valid(context.Background(), 1, now) {
		t.Fatal("a token issued in the same second as the revocation must be rejected")
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
