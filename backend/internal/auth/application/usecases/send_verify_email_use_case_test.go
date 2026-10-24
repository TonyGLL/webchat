package usecases

import "testing"

func TestBuildVerifyURL(t *testing.T) {
	tests := []struct {
		name, base, token, want string
	}{
		{"plain", "http://localhost:3100", "abc.def", "http://localhost:3100/verify/abc.def"},
		{"trailing slash", "https://chat.example.com/", "abc", "https://chat.example.com/verify/abc"},
		{"escapes token", "http://localhost:3100", "a/b c", "http://localhost:3100/verify/a%2Fb%20c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildVerifyURL(tt.base, tt.token); got != tt.want {
				t.Fatalf("buildVerifyURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
