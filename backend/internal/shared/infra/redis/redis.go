package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

// Redis sits on the per-request path (session guard, rate limiter, tickets). The go-redis
// defaults (3 s reads, 3 retries, 4 s pool wait) would let a hung Redis hold every request for
// ~12 s before the fail-open paths kick in, so the socket and retry budget is much tighter.
const (
	dialTimeout  = time.Second
	ioTimeout    = 500 * time.Millisecond
	poolTimeout  = time.Second
	maxRetries   = 1
	minRetryWait = 8 * time.Millisecond
)

func tune(opts *redis.Options) {
	opts.DialTimeout = dialTimeout
	opts.ReadTimeout = ioTimeout
	opts.WriteTimeout = ioTimeout
	opts.PoolTimeout = poolTimeout
	opts.MaxRetries = maxRetries
	opts.MinRetryBackoff = minRetryWait
}

// NewClient establishes a connection to a Redis server using the provided URL.
// It pings the server to verify the connection is active before returning the client.
func NewClient(ctx context.Context, redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse redis URL: %w", err)
	}

	// go-redis v9 only looks at a context's deadline or cancellation for socket I/O when this is
	// on (v8 always did). The hub bounds presence calls with a 2 s context so a stalled Redis
	// cannot hold up connection cleanup (D24), and request handlers rely on their context being
	// cancelled when the client goes away: without the option those bounds would silently stop
	// working.
	opts.ContextTimeoutEnabled = true
	tune(opts)
	// One standalone Redis, no failover or cluster handoffs to follow: skip the extra
	// CLIENT MAINT_NOTIFICATIONS handshake and the background workers that "auto" mode adds.
	opts.MaintNotificationsConfig = &maintnotifications.Config{Mode: maintnotifications.ModeDisabled}

	client := redis.NewClient(opts)

	// Ping the Redis server to ensure a connection is established.
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("could not connect to redis: %w", err)
	}

	return client, nil
}
