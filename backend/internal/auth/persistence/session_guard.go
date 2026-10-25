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
	// Backend lookups run on every authenticated request and fail open, so they get short bounds.
	revocationTimeout  = 250 * time.Millisecond
	activeCheckTimeout = time.Second
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

	mu        sync.Mutex
	cache     map[int]activeEntry
	lastSweep time.Time
}

func NewSessionGuard(tokens domain.TokenRepository, users ActiveChecker) *SessionGuard {
	return &SessionGuard{tokens: tokens, users: users, now: time.Now, cache: map[int]activeEntry{}}
}

func (g *SessionGuard) Valid(ctx context.Context, userID int, issuedAt time.Time) bool {
	rctx, cancel := context.WithTimeout(ctx, revocationTimeout)
	revokedAt, err := g.tokens.SessionsRevokedAt(rctx, userID)
	cancel()
	if err != nil {
		log.Printf("session guard: revocation lookup failed, allowing: %v", err)
	} else if !revokedAt.IsZero() && !issuedAt.After(revokedAt) {
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

	actx, cancel := context.WithTimeout(ctx, activeCheckTimeout)
	active, err := g.users.IsActive(actx, userID)
	cancel()
	if err != nil {
		log.Printf("session guard: account lookup failed, allowing: %v", err)
		return true
	}

	g.mu.Lock()
	g.makeRoomLocked(now)
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

// makeRoomLocked keeps the cache below activeCacheMax without throwing everything away (a wipe
// would send every request to the database at once when more than activeCacheMax users are
// active): it drops the expired entries (at most once per TTL, the scan is O(n)) and, if the
// cache is still full of live entries, one arbitrary entry (Go's map iteration order is random).
func (g *SessionGuard) makeRoomLocked(now time.Time) {
	if len(g.cache) < activeCacheMax {
		return
	}
	if now.Sub(g.lastSweep) >= activeCacheTTL {
		g.lastSweep = now
		for id, e := range g.cache {
			if !now.Before(e.expires) {
				delete(g.cache, id)
			}
		}
	}
	for id := range g.cache {
		if len(g.cache) < activeCacheMax {
			return
		}
		delete(g.cache, id)
	}
}
