package http

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"backend/internal/shared/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// Server holds the dependencies for the HTTP server.
type Server struct {
	config *config.Config
	Router *gin.Engine
}

// NewServer creates a new HTTP server and sets up shared middleware.
func NewServer(config *config.Config) *Server {
	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(RequestLogger(), SafeRecovery())
	server := &Server{
		config: config,
		Router: router,
	}
	// By default trust no proxy headers: the client IP is the TCP peer, so X-Forwarded-For
	// cannot be spoofed to dodge rate limits. Behind a load balancer set TRUSTED_PROXIES, or
	// every user would share one rate-limit bucket.
	var trusted []string
	if len(config.TrustedProxies) > 0 {
		trusted = config.TrustedProxies
	}
	if err := router.SetTrustedProxies(trusted); err != nil {
		log.Fatalf("invalid TRUSTED_PROXIES: %v", err)
	}
	server.setupMiddleware()
	return server
}

func (s *Server) setupMiddleware() {
	s.Router.Use(cors.New(cors.Config{
		AllowOrigins:     s.config.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		// Browsers cap this (Chrome 2 h, Firefox 24 h); it saves one OPTIONS round trip before
		// every cross-origin call with a JSON body or an Authorization header.
		MaxAge: corsMaxAge,
	}))
	// Compress API JSON (history pages, member lists, the directory). The websocket upgrade is
	// skipped by the middleware itself and /ws by path; small bodies are not worth the CPU.
	s.Router.Use(gzip.Gzip(gzip.DefaultCompression,
		gzip.WithExcludedPaths([]string{"/api/v1/ws"}),
		gzip.WithMinLength(1024),
	))
	s.Router.Use(APIHeaders())
	// After CORS so a 413 still carries the CORS headers the browser needs to read it.
	s.Router.Use(MaxBodySize(DefaultMaxBodyBytes))
}

const corsMaxAge = 2 * time.Hour

// APIHeaders sets `X-Content-Type-Options: nosniff` on every response and
// `Cache-Control: no-store` on the API (/api/...): token bodies and authenticated JSON must not
// be kept by browser or proxy caches.
func APIHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			h.Set("Cache-Control", "no-store")
		}
		c.Next()
	}
}

// Limits on the underlying http.Server. Hijacked connections (the websocket at /ws) are not
// subject to them: gorilla/websocket clears the deadlines the server set when it upgrades.
const (
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	writeTimeout      = 60 * time.Second
	idleTimeout       = 120 * time.Second
	maxHeaderBytes    = 1 << 16
	// shutdownTimeout bounds the whole stop sequence (shutdown hooks, then draining requests).
	shutdownTimeout = 15 * time.Second
)

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}
}

// Run serves HTTP until ctx is cancelled (main ties it to SIGINT/SIGTERM) and then stops in order:
// the onShutdown hooks run first, while the listener still answers (e.g. the websocket hub sends
// its close frames), then the server drains in-flight requests. It returns the listen error, or
// the shutdown error if draining timed out; it never exits the process, so the caller's deferred
// cleanups (database, Redis) still run.
func (s *Server) Run(ctx context.Context, onShutdown ...func(context.Context)) error {
	ln, err := net.Listen("tcp", ":"+s.config.Port)
	if err != nil {
		return err
	}
	log.Printf("Server running on port: %s", s.config.Port)
	return serve(ctx, newHTTPServer(ln.Addr().String(), s.Router), ln, onShutdown)
}

func serve(ctx context.Context, srv *http.Server, ln net.Listener, onShutdown []func(context.Context)) error {
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
	}
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	for _, hook := range onShutdown {
		hook(shutdownCtx)
	}
	if err := srv.Shutdown(shutdownCtx); err != nil {
		_ = srv.Close()
		return err
	}
	log.Println("Server exiting")
	return nil
}
