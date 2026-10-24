package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type fakeCounter struct{ hits map[string]int64 }

func (f *fakeCounter) Hit(_ context.Context, key string, window time.Duration) (int64, time.Duration, error) {
	f.hits[key]++
	return f.hits[key], window, nil
}

func newTestRouter(counter Counter) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.Use(RateLimit(counter, []Rule{{Method: "POST", Path: "/login", Limit: 2, Window: time.Minute}}))
	r.POST("/login", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/free", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func do(r *gin.Engine, method, path, remote string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.RemoteAddr = remote
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRateLimitBlocksAfterLimitPerIP(t *testing.T) {
	r := newTestRouter(&fakeCounter{hits: map[string]int64{}})

	for i := 1; i <= 2; i++ {
		if w := do(r, "POST", "/login", "1.1.1.1:1000"); w.Code != http.StatusOK {
			t.Fatalf("request %d: status %d, want 200", i, w.Code)
		}
	}
	w := do(r, "POST", "/login", "1.1.1.1:1000")
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("third request: status %d, want 429", w.Code)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Fatal("missing Retry-After header")
	}
	if w := do(r, "POST", "/login", "2.2.2.2:1000"); w.Code != http.StatusOK {
		t.Fatalf("other IP: status %d, want 200", w.Code)
	}
}

func TestRateLimitIgnoresUnlistedRoutes(t *testing.T) {
	r := newTestRouter(&fakeCounter{hits: map[string]int64{}})
	for i := 0; i < 10; i++ {
		if w := do(r, "GET", "/free", "1.1.1.1:1000"); w.Code != http.StatusOK {
			t.Fatalf("status %d, want 200", w.Code)
		}
	}
}
