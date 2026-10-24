package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Pinger is anything that can report whether it is reachable (database, Redis...).
type Pinger interface {
	Ping(ctx context.Context) error
}

// PingFunc adapts a function to the Pinger interface.
type PingFunc func(ctx context.Context) error

func (f PingFunc) Ping(ctx context.Context) error { return f(ctx) }

// RegisterHealthRoutes adds:
//   - GET /health: the process is up (liveness)
//   - GET /ready: dependencies answer (readiness); 503 if any does not
func RegisterHealthRoutes(router *gin.Engine, deps map[string]Pinger) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		checks := make(map[string]string, len(deps))
		ready := true
		for name, dep := range deps {
			if err := dep.Ping(ctx); err != nil {
				checks[name] = "unavailable"
				ready = false
			} else {
				checks[name] = "ok"
			}
		}

		status, label := http.StatusOK, "ready"
		if !ready {
			status, label = http.StatusServiceUnavailable, "not ready"
		}
		c.JSON(status, gin.H{"status": label, "checks": checks})
	})
}
