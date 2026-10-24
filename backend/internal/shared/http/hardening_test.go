package http

import (
	"bytes"
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func TestMaxBodySizeRejectsDeclaredOversizeWith413JSON(t *testing.T) {
	r := gin.New()
	r.Use(MaxBodySize(10))
	called := false
	r.POST("/x", func(c *gin.Context) { called = true; c.Status(http.StatusOK) })

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(strings.Repeat("a", 11))))

	if rec.Code != http.StatusRequestEntityTooLarge || called {
		t.Fatalf("status=%d called=%v, want 413 and handler not run", rec.Code, called)
	}
	if !strings.Contains(rec.Body.String(), `"error"`) {
		t.Fatalf("body should be the JSON error shape, got %q", rec.Body.String())
	}
}

func TestMaxBodySizeCutsUndeclaredBodiesAndAllowsSmallOnes(t *testing.T) {
	r := gin.New()
	r.Use(MaxBodySize(10))
	r.POST("/x", func(c *gin.Context) {
		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		c.String(http.StatusOK, string(b))
	})

	small := httptest.NewRecorder()
	r.ServeHTTP(small, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader("hello")))
	if small.Code != http.StatusOK || small.Body.String() != "hello" {
		t.Fatalf("small body: %d %q", small.Code, small.Body.String())
	}

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(strings.Repeat("a", 50)))
	req.ContentLength = -1 // chunked: no declared length
	big := httptest.NewRecorder()
	r.ServeHTTP(big, req)
	if big.Code != http.StatusBadRequest {
		t.Fatalf("undeclared oversize body must fail to read, got %d", big.Code)
	}
}

func TestRequestLoggerReplacesBadRequestIDs(t *testing.T) {
	r := gin.New()
	r.Use(RequestLogger())
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	cases := []struct {
		in   string
		keep bool
	}{
		{"abc-123_X.y", true},
		{strings.Repeat("a", 64), true},
		{strings.Repeat("a", 65), false},
		{"has space", false},
		{"<script>alert(1)</script>", false},
		{"", false},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		if tc.in != "" {
			req.Header.Set(requestIDHeader, tc.in)
		}
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		got := rec.Header().Get(requestIDHeader)
		if tc.keep && got != tc.in {
			t.Errorf("%q should be kept, got %q", tc.in, got)
		}
		if !tc.keep && (got == tc.in || !validRequestID.MatchString(got)) {
			t.Errorf("%q should be replaced by a valid ID, got %q", tc.in, got)
		}
	}
}

func TestSafeRecoveryReturns500WithoutLeakingHeaders(t *testing.T) {
	var logs bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(prev)

	r := gin.New()
	r.Use(SafeRecovery())
	r.GET("/boom", func(c *gin.Context) { panic("kaboom") })

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	req.Header.Set("Cookie", "refresh_token=SECRET")
	req.Header.Set("Authorization", "Bearer SECRET")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), `"error"`) {
		t.Fatalf("got %d %q", rec.Code, rec.Body.String())
	}
	if strings.Contains(logs.String(), "SECRET") || strings.Contains(rec.Body.String(), "SECRET") {
		t.Fatal("credentials leaked")
	}
}

// Shutdown order: the hooks must run while the server still answers, and the server must be
// closed once serve returns.
func TestServeRunsShutdownHooksBeforeClosingTheServer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	srv := newHTTPServer(ln.Addr().String(), mux)
	url := "http://" + ln.Addr().String() + "/ping"

	var mu sync.Mutex
	var order []string
	hook := func(context.Context) {
		resp, err := http.Get(url)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			order = append(order, "hook:server-down")
			return
		}
		resp.Body.Close()
		order = append(order, "hook:server-up")
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- serve(ctx, srv, ln, []func(context.Context){hook}) }()

	// Wait until it serves, then trigger the shutdown.
	for i := 0; ; i++ {
		if resp, err := http.Get(url); err == nil {
			resp.Body.Close()
			break
		}
		if i > 100 {
			t.Fatal("server did not start")
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("serve: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not return")
	}
	mu.Lock()
	defer mu.Unlock()
	if len(order) != 1 || order[0] != "hook:server-up" {
		t.Fatalf("hook order = %v, want it to run while the server is up", order)
	}
	if resp, err := http.Get(url); err == nil {
		resp.Body.Close()
		t.Fatal("server must be closed after serve returns")
	}
}
