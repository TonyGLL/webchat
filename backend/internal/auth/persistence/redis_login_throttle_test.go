package persistence

import (
	"context"
	"os"
	"strings"
	"testing"

	"backend/internal/auth/domain"

	"github.com/redis/go-redis/v9"
)

// Needs a real Redis: set TEST_REDIS_ADDR (e.g. localhost:57380).
func TestRedisLoginThrottle(t *testing.T) {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR not set")
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close()
	ctx := context.Background()
	repo := NewRedisTokenRepository(client).(domain.LoginThrottle)
	id := "Ana-" + t.Name()
	t.Cleanup(func() {
		a, p := throttleKeys(id, "1.1.1.1")
		_, p2 := throttleKeys(id, "2.2.2.2")
		for _, k := range []string{a, p, p2} {
			client.Del(ctx, "login:fail:"+k, "login:wait:"+k)
		}
	})

	for i := 0; i < domain.ThrottlePairThreshold-1; i++ {
		_ = repo.RecordFailure(ctx, id, "1.1.1.1")
	}
	if wait, err := repo.Check(ctx, id, "1.1.1.1"); err != nil || wait != 0 {
		t.Fatalf("below the threshold: wait=%v err=%v", wait, err)
	}
	_ = repo.RecordFailure(ctx, id, "1.1.1.1")
	if wait, err := repo.Check(ctx, id, "1.1.1.1"); err != nil || wait <= 0 {
		t.Fatalf("at the threshold: wait=%v err=%v", wait, err)
	}
	// the identifier is case-insensitive, another address is unaffected
	if wait, _ := repo.Check(ctx, strings.ToUpper(id), "1.1.1.1"); wait <= 0 {
		t.Fatal("identifiers must match regardless of case")
	}
	if wait, _ := repo.Check(ctx, id, "2.2.2.2"); wait != 0 {
		t.Fatalf("other address must not wait, got %v", wait)
	}
	if err := repo.RecordSuccess(ctx, id, "1.1.1.1"); err != nil {
		t.Fatal(err)
	}
	if wait, _ := repo.Check(ctx, id, "1.1.1.1"); wait != 0 {
		t.Fatalf("a success must clear the pair, got %v", wait)
	}
}
