package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

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
