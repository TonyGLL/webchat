package http

import (
	"net/http"
	"testing"
)

func TestClientKeyGroupsIPv6By64(t *testing.T) {
	tests := []struct{ in, want string }{
		{"203.0.113.7", "203.0.113.7"},
		{"2001:db8:1:2:aaaa:bbbb:cccc:dddd", "2001:db8:1:2::/64"},
		{"2001:db8:1:2::1", "2001:db8:1:2::/64"},
		{"2001:db8:1:3::1", "2001:db8:1:3::/64"},
		{"::ffff:203.0.113.7", "203.0.113.7"},
		{"not-an-ip", "not-an-ip"},
	}
	for _, tt := range tests {
		if got := clientKey(tt.in); got != tt.want {
			t.Errorf("clientKey(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// Rotating addresses inside one /64 must share one bucket.
func TestRateLimitSharesBucketInsideIPv6Prefix(t *testing.T) {
	r := newTestRouter(&fakeCounter{hits: map[string]int64{}})
	for i, remote := range []string{"[2001:db8:1:2::1]:1", "[2001:db8:1:2::2]:1"} {
		if w := do(r, "POST", "/login", remote); w.Code != http.StatusOK {
			t.Fatalf("request %d: status %d", i, w.Code)
		}
	}
	if w := do(r, "POST", "/login", "[2001:db8:1:2:ffff::9]:1"); w.Code != http.StatusTooManyRequests {
		t.Fatalf("third address in the same /64: status %d, want 429", w.Code)
	}
	if w := do(r, "POST", "/login", "[2001:db8:1:99::1]:1"); w.Code != http.StatusOK {
		t.Fatalf("other /64: status %d, want 200", w.Code)
	}
}
