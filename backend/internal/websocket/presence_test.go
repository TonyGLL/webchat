package websocket

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestConnectionsOf(t *testing.T) {
	members := []string{"1:a", "1:b", "12:c", "2:d"}
	if got := connectionsOf(members, 1); got != 2 {
		t.Fatalf("user 1 = %d, want 2 (user 12 must not match)", got)
	}
	if got := connectionsOf(members, 3); got != 0 {
		t.Fatalf("user 3 = %d, want 0", got)
	}
}

// Needs a real Redis: set TEST_REDIS_ADDR (e.g. localhost:6380) to run it.
func TestRedisPresenceLifecycle(t *testing.T) {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR not set")
	}
	client := redis.NewClient(&redis.Options{Addr: addr, Password: os.Getenv("TEST_REDIS_PASSWORD")})
	ctx := context.Background()
	client.Del(ctx, presenceKey)
	t.Cleanup(func() { client.Del(ctx, presenceKey); client.Close() })

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
}
