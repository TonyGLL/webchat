package presentation

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/internal/auth/application/usecases"
	"backend/internal/auth/domain"
	shared_domain "backend/internal/shared/domain"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type failingTokens struct {
	domain.TokenRepository
	err error
}

func (f failingTokens) GetToken(context.Context, string, string) (int, error) { return 0, f.err }

func verifyRequest(t *testing.T, tokenErr error) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	uc := usecases.NewVerifyEmailUseCase(nil, failingTokens{err: tokenErr}, nil, nil)
	ctrl := &AuthController{VerifyEmailUseCase: uc, Validate: validator.New()}
	r := gin.New()
	r.POST("/auth/verify-email/:token", ctrl.VerifyEmail)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/verify-email/abc", nil))
	return w
}

// An unknown token used to surface as a 500 with the internal error text.
func TestVerifyEmailUnknownTokenIs400(t *testing.T) {
	w := verifyRequest(t, shared_domain.ErrNotFound)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", w.Code)
	}
}

func TestVerifyEmailBackendFailureDoesNotLeakInternalError(t *testing.T) {
	w := verifyRequest(t, errors.New("dial tcp 10.0.0.5:6379: connect: connection refused"))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status %d, want 500", w.Code)
	}
	if strings.Contains(w.Body.String(), "10.0.0.5") || strings.Contains(w.Body.String(), "dial") {
		t.Fatalf("500 body leaks the internal error: %s", w.Body.String())
	}
}

func TestLoginValidationErrorDoesNotLeakStructNames(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := &AuthController{Validate: validator.New()}
	r := gin.New()
	r.POST("/auth/login", ctrl.Login)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"identifier":"","password":""}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", w.Code)
	}
	if strings.Contains(w.Body.String(), "Key:") || strings.Contains(w.Body.String(), "DTO") {
		t.Fatalf("400 body leaks validator text: %s", w.Body.String())
	}
}
