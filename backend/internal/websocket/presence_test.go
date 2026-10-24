package websocket

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func testRedis(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR not set")
	}
	client := redis.NewClient(&redis.Options{Addr: addr, Password: os.Getenv("TEST_REDIS_PASSWORD")})
	t.Cleanup(func() { client.Close() })
	return client
}

// Needs a real Redis: set TEST_REDIS_ADDR (e.g. localhost:6380) to run it.
func TestRedisPresenceLifecycle(t *testing.T) {
	client := testRedis(t)
	ctx := context.Background()
	clean := func() { client.Del(ctx, userKey(5), userKey(7), userKey(8), userKey(9)) }
	clean()
	t.Cleanup(clean)

	p := NewRedisPresence(client)
	if first, err := p.Connect(ctx, 7, "tab1"); err != nil || !first {
		t.Fatalf("first connect: first=%v err=%v", first, err)
	}
	if first, _ := p.Connect(ctx, 7, "tab2"); first {
		t.Fatal("second connection must not count as first")
	}
	online, _ := p.OnlineUsers(ctx, []int{7, 8})
	if !online[7] || online[8] {
		t.Fatalf("online = %v", online)
	}
	if last, _ := p.Disconnect(ctx, 7, "tab1"); last {
		t.Fatal("one connection left, not last")
	}
	if last, _ := p.Disconnect(ctx, 7, "tab2"); !last {
		t.Fatal("last connection should report last")
	}

	// a connection that stops heartbeating expires by itself
	p.now = func() time.Time { return time.Now().Add(-presenceTTL - time.Minute) }
	_, _ = p.Connect(ctx, 9, "ghost")
	p.now = time.Now
	if online, _ := p.OnlineUsers(ctx, []int{9}); online[9] {
		t.Fatal("expired connection must not count as online")
	}

	// the per-user key carries a TTL so abandoned users disappear by themselves
	_, _ = p.Connect(ctx, 8, "tab")
	if ttl := client.TTL(ctx, userKey(8)).Val(); ttl <= 0 || ttl > presenceTTL {
		t.Fatalf("user key ttl = %v", ttl)
	}
	// a stale sibling is pruned when the user connects again: still "first"
	p.now = func() time.Time { return time.Now().Add(-presenceTTL - time.Minute) }
	_, _ = p.Connect(ctx, 5, "stale")
	p.now = time.Now
	if first, _ := p.Connect(ctx, 5, "fresh"); !first {
		t.Fatal("an expired connection must not make the new one a non-first")
	}
}

func TestRedisPresenceLookupIsCapped(t *testing.T) {
	client := testRedis(t)
	ctx := context.Background()
	p := NewRedisPresence(client)
	ids := make([]int, maxOnlineLookup+50)
	for i := range ids {
		ids[i] = 900000 + i
	}
	first, last := ids[0], ids[len(ids)-1]
	t.Cleanup(func() { client.Del(ctx, userKey(first), userKey(last)) })
	_, _ = p.Connect(ctx, first, "x")
	_, _ = p.Connect(ctx, last, "x") // beyond the cap
	online, err := p.OnlineUsers(ctx, ids)
	if err != nil || !online[first] || online[last] {
		t.Fatalf("online=%d err=%v", len(online), err)
	}
}
