package http

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/shared/infra/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

type fakeGuard struct {
	ok     bool
	userID int
	iat    time.Time
}

func (g *fakeGuard) Valid(_ context.Context, userID int, iat time.Time) bool {
	g.userID, g.iat = userID, iat
	return g.ok
}

func call(t *testing.T, token string, guard SessionGuard) int {
	t.Helper()
	gin.SetMode(gin.TestMode)
	svc, err := services.NewJWTService(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.GET("/", JWTMiddleware(svc, guard), func(c *gin.Context) { c.Status(nethttp.StatusOK) })
	req := httptest.NewRequest(nethttp.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestMiddlewareOnlyAcceptsAccessTokens(t *testing.T) {
	svc, _ := services.NewJWTService(testSecret)

	access, _ := svc.GenerateToken(7, time.Minute)
	refresh, _ := svc.GenerateRefreshToken(7, "jti", time.Minute)
	verify, _ := svc.GenerateTokenFromClaims(map[string]interface{}{"user_id": 7, "email": "a@b.c", "typ": "verify"}, time.Minute)
	noType, _ := svc.GenerateTokenFromClaims(map[string]interface{}{"user_id": 7}, time.Minute)

	// A token minted before the "typ" claim existed: valid signature, no typ, no jti.
	legacy, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 7, "iat": time.Now().Unix(), "exp": time.Now().Add(time.Minute).Unix(),
	}).SignedString([]byte(testSecret))

	hs512, _ := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"user_id": 7, "typ": "access", "exp": time.Now().Add(time.Minute).Unix(),
	}).SignedString([]byte(testSecret))

	tests := []struct {
		name  string
		token string
		want  int
	}{
		{"access", access, nethttp.StatusOK},
		{"refresh", refresh, nethttp.StatusUnauthorized},
		{"verify", verify, nethttp.StatusUnauthorized},
		{"claims without a type", noType, nethttp.StatusUnauthorized}, // generic generator defaults to verify
		{"legacy token without typ", legacy, nethttp.StatusUnauthorized},
		{"HS512", hs512, nethttp.StatusUnauthorized},
		{"garbage", "nope", nethttp.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := call(t, tt.token, nil); got != tt.want {
				t.Fatalf("status = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMiddlewareConsultsTheSessionGuard(t *testing.T) {
	svc, _ := services.NewJWTService(testSecret)
	access, _ := svc.GenerateToken(7, time.Minute)

	guard := &fakeGuard{ok: false}
	if got := call(t, access, guard); got != nethttp.StatusUnauthorized {
		t.Fatalf("a revoked session must be rejected, got %d", got)
	}
	if guard.userID != 7 || guard.iat.IsZero() {
		t.Fatalf("guard got user %d, iat %v", guard.userID, guard.iat)
	}

	guard.ok = true
	if got := call(t, access, guard); got != nethttp.StatusOK {
		t.Fatalf("a live session must pass, got %d", got)
	}
}
