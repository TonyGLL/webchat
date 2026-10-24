package persistence

import (
	"context"
	"log"
	"sync"
	"time"

	"backend/internal/auth/domain"
)

// ActiveChecker tells whether an account exists and is not deactivated.
type ActiveChecker interface {
	IsActive(ctx context.Context, userID int) (bool, error)
}

// activeCacheTTL bounds how long "this account is active" is trusted. Deactivation and password
// resets also write a session revocation to Redis, which is checked on every request, so in
// practice they take effect at once; the cache only covers an account flagged some other way.
const (
	activeCacheTTL = 15 * time.Second
	activeCacheMax = 10000
)

type activeEntry struct {
	active  bool
	expires time.Time
}

// SessionGuard implements the HTTP middleware's SessionGuard. It rejects tokens issued before
// the user's sessions were revoked and tokens of deleted or deactivated accounts. Backend
// failures (Redis, Postgres) fail open and are logged: a short outage must not sign everybody out.
type SessionGuard struct {
	tokens domain.TokenRepository
	users  ActiveChecker
	now    func() time.Time

	mu    sync.Mutex
	cache map[int]activeEntry
}

func NewSessionGuard(tokens domain.TokenRepository, users ActiveChecker) *SessionGuard {
	return &SessionGuard{tokens: tokens, users: users, now: time.Now, cache: map[int]activeEntry{}}
}

func (g *SessionGuard) Valid(ctx context.Context, userID int, issuedAt time.Time) bool {
	revokedAt, err := g.tokens.SessionsRevokedAt(ctx, userID)
	if err != nil {
		log.Printf("session guard: revocation lookup failed, allowing: %v", err)
	} else if !revokedAt.IsZero() && issuedAt.Before(revokedAt) {
		return false
	}
	return g.isActive(ctx, userID)
}

func (g *SessionGuard) isActive(ctx context.Context, userID int) bool {
	now := g.now()
	g.mu.Lock()
	entry, ok := g.cache[userID]
	g.mu.Unlock()
	if ok && now.Before(entry.expires) {
		return entry.active
	}

	active, err := g.users.IsActive(ctx, userID)
	if err != nil {
		log.Printf("session guard: account lookup failed, allowing: %v", err)
		return true
	}

	g.mu.Lock()
	if len(g.cache) >= activeCacheMax {
		g.cache = map[int]activeEntry{}
	}
	g.cache[userID] = activeEntry{active: active, expires: now.Add(activeCacheTTL)}
	g.mu.Unlock()
	return active
}

// Invalidate forgets the cached state of one account.
func (g *SessionGuard) Invalidate(userID int) {
	g.mu.Lock()
	delete(g.cache, userID)
	g.mu.Unlock()
}
