package http

import (
	"log/slog"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

// validRequestID limits what a client may put in X-Request-ID: it is echoed in a response header
// and written to the logs, so anything else (newlines, huge values) is replaced.
var validRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// quietPaths are probes hit every few seconds by orchestrators; they are logged at debug level only.
var quietPaths = map[string]struct{}{"/health": {}, "/ready": {}}

// RequestLogger logs one structured line per request and tags it with a request ID,
// which is also returned in the X-Request-ID response header for correlation.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := c.GetHeader(requestIDHeader)
		if !validRequestID.MatchString(requestID) {
			requestID = uuid.NewString()
		}
		c.Header(requestIDHeader, requestID)

		c.Next()

		status := c.Writer.Status()
		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}
		if _, quiet := quietPaths[c.FullPath()]; quiet && status < 400 {
			level = slog.LevelDebug
		}

		slog.Log(c.Request.Context(), level, "request",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
