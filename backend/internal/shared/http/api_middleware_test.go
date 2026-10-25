package http

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/internal/shared/config"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func newAPITestServer() *Server {
	s := NewServer(&config.Config{CORSAllowedOrigins: []string{"http://localhost:3100"}})
	s.Router.GET("/api/v1/big", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": strings.Repeat("x", 4000)})
	})
	s.Router.GET("/api/v1/small", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	s.Router.GET("/docs", func(c *gin.Context) { c.String(http.StatusOK, "docs") })
	up := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	s.Router.GET("/api/v1/ws", func(c *gin.Context) {
		conn, err := up.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.WriteMessage(websocket.TextMessage, []byte("hello"))
	})
	return s
}

func TestCORSPreflightCarriesMaxAge(t *testing.T) {
	r := newAPITestServer().Router
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/small", nil)
	req.Header.Set("Origin", "http://localhost:3100")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if got := w.Header().Get("Access-Control-Max-Age"); got != "7200" {
		t.Fatalf("Access-Control-Max-Age = %q, want 7200", got)
	}
}

func TestAPIHeaders(t *testing.T) {
	r := newAPITestServer().Router
	for path, wantCache := range map[string]string{"/api/v1/small": "no-store", "/docs": ""} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("%s: X-Content-Type-Options = %q", path, got)
		}
		if got := w.Header().Get("Cache-Control"); got != wantCache {
			t.Errorf("%s: Cache-Control = %q, want %q", path, got, wantCache)
		}
	}
}

func TestGzipCompressesLargeJSONOnly(t *testing.T) {
	r := newAPITestServer().Router

	req := httptest.NewRequest(http.MethodGet, "/api/v1/big", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("large JSON not compressed: %v", w.Header())
	}
	if w.Body.Len() > 1000 {
		t.Fatalf("compressed body is %d bytes, expected far less than 4000", w.Body.Len())
	}
	zr, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(zr)
	if !strings.Contains(string(body), strings.Repeat("x", 4000)) {
		t.Fatal("decompressed body is wrong")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/small", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Header().Get("Content-Encoding") != "" {
		t.Fatal("tiny body should not be compressed")
	}
}

// The websocket upgrade must pass through the compression middleware untouched.
func TestGzipLeavesWebsocketUpgradeAlone(t *testing.T) {
	srv := httptest.NewServer(newAPITestServer().Router)
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/v1/ws"
	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(url, http.Header{"Accept-Encoding": []string{"gzip"}})
	if err != nil {
		t.Fatalf("websocket dial through the middleware stack: %v", err)
	}
	defer conn.Close()
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("status %d", resp.StatusCode)
	}
	_, msg, err := conn.ReadMessage()
	if err != nil || string(msg) != "hello" {
		t.Fatalf("read: %q, %v", msg, err)
	}
}
