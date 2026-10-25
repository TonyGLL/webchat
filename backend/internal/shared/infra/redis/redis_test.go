package redis

import (
	"context"
	"net"
	"testing"
	"time"
)

// A server that accepts connections and never answers stands in for a stalled Redis. NewClient
// must give up when the caller's context does, not after go-redis's own read timeout (5 s):
// the hub and the HTTP handlers bound their Redis calls with contexts (D24), and go-redis v9
// ignores them unless ContextTimeoutEnabled is set.
func TestNewClientHonoursContextDeadline(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var conns []net.Conn
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			conns = append(conns, c)
		}
	}()
	t.Cleanup(func() {
		ln.Close()
		<-done
		for _, c := range conns {
			c.Close()
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := NewClient(ctx, "redis://"+ln.Addr().String()); err == nil {
		t.Fatal("expected an error from a server that never answers")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("NewClient took %v: the context deadline was ignored", elapsed)
	}
}
