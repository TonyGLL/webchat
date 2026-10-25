package http

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// Needs a real Redis (6.2 or later): set TEST_REDIS_ADDR. The old EXPIRE NX needed 7.0 and made
// the limiter silently fail open on older servers.
func TestRedisCounterHit(t *testing.T) {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR not set")
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close()
	ctx := context.Background()
	key := "rl:test:" + time.Now().Format(time.RFC3339Nano)
	t.Cleanup(func() { client.Del(ctx, key) })

	c := NewRedisCounter(client)
	for want := int64(1); want <= 3; want++ {
		n, ttl, err := c.Hit(ctx, key, time.Minute)
		if err != nil {
			t.Fatalf("hit %d: %v", want, err)
		}
		if n != want {
			t.Fatalf("count = %d, want %d", n, want)
		}
		if ttl <= 0 || ttl > time.Minute {
			t.Fatalf("ttl = %v, want within (0, 1m]", ttl)
		}
	}

	// A key left without a TTL (crash between INCR and EXPIRE in the old scheme) heals.
	orphan := key + ":orphan"
	t.Cleanup(func() { client.Del(ctx, orphan) })
	client.Set(ctx, orphan, 5, 0)
	n, ttl, err := c.Hit(ctx, orphan, time.Minute)
	if err != nil || n != 6 || ttl <= 0 {
		t.Fatalf("orphan key: n=%d ttl=%v err=%v", n, ttl, err)
	}
}
