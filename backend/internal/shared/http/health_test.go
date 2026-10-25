package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthAndReady(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterHealthRoutes(router, map[string]Pinger{
		"db":    PingFunc(func(context.Context) error { return nil }),
		"redis": PingFunc(func(context.Context) error { return errors.New("down") }),
	})

	get := func(path string) int {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		return w.Code
	}

	if code := get("/health"); code != http.StatusOK {
		t.Fatalf("/health = %d, want 200", code)
	}
	if code := get("/ready"); code != http.StatusServiceUnavailable {
		t.Fatalf("/ready with a dead dependency = %d, want 503", code)
	}
}

func TestReadyFlipsOnShutdown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	markDraining := RegisterHealthRoutes(router, map[string]Pinger{
		"db": PingFunc(func(context.Context) error { return nil }),
	})
	get := func(path string) int {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		return w.Code
	}
	if code := get("/ready"); code != http.StatusOK {
		t.Fatalf("/ready before shutdown = %d, want 200", code)
	}
	markDraining(context.Background())
	if code := get("/ready"); code != http.StatusServiceUnavailable {
		t.Fatalf("/ready after shutdown starts = %d, want 503", code)
	}
	if code := get("/health"); code != http.StatusOK {
		t.Fatalf("/health must stay 200 while draining, got %d", code)
	}
}
