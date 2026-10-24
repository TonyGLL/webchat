package websocket

import (
	"net/http"
	"testing"
)

func TestOriginChecker(t *testing.T) {
	tests := []struct {
		name    string
		allowed []string
		origin  string
		host    string
		want    bool
	}{
		{"no origin (non-browser)", []string{"http://localhost:3100"}, "", "api:8081", true},
		{"allowed origin", []string{"http://localhost:3100"}, "http://localhost:3100", "api:8081", true},
		{"allowed with trailing slash in config", []string{"http://localhost:3100/"}, "http://localhost:3100", "api:8081", true},
		{"foreign origin", []string{"http://localhost:3100"}, "https://evil.example", "api:8081", false},
		{"empty list, same origin", nil, "http://api:8081", "api:8081", true},
		{"empty list, foreign origin", nil, "https://evil.example", "api:8081", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{Host: tt.host, Header: http.Header{}}
			if tt.origin != "" {
				r.Header.Set("Origin", tt.origin)
			}
			if got := originChecker(tt.allowed)(r); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
