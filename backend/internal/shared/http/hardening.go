package http

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// DefaultMaxBodyBytes is the largest request body the API accepts (messages are capped far below it).
const DefaultMaxBodyBytes int64 = 1 << 20 // 1 MiB

// MaxBodySize rejects request bodies larger than limit bytes. A declared Content-Length over the
// limit gets a 413 JSON error at once; bodies without a length (chunked) are cut by
// http.MaxBytesReader while the handler reads them, which surfaces as a bind error.
func MaxBodySize(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > limit {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request body too large"})
			return
		}
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		}
		c.Next()
	}
}

// SafeRecovery turns a panic into a 500 JSON error. Unlike gin.Recovery it never dumps the request
// headers (Cookie and Authorization carry credentials): only the method, route and stack are logged.
func SafeRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if err, ok := rec.(error); ok && errors.Is(err, http.ErrAbortHandler) {
				// The client went away mid-response: nothing to report, just stop.
				c.Abort()
				return
			}
			slog.Error("panic recovered",
				"panic", rec,
				"method", c.Request.Method,
				"path", c.FullPath(),
				"stack", string(debug.Stack()),
			)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}()
		c.Next()
	}
}
