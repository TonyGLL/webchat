package usecases

import (
	"context"
	"testing"
	"time"

	"backend/internal/auth/application/dtos"
	"backend/internal/auth/domain"
	shared_app "backend/internal/shared/application"
	shared_domain "backend/internal/shared/domain"

	"github.com/golang-jwt/jwt/v5"
)

type fakeAuthRepo struct{ domain.AuthRepository }

func (fakeAuthRepo) GetUserByID(_ context.Context, id int) (*domain.User, error) {
	return &domain.User{ID: id, Username: "ana"}, nil
}

// fakeTokens keeps tokens with an absolute expiry so tests can move the clock.
type fakeTokens struct {
	now     time.Time
	expires map[string]time.Time
	users   map[string]int
}

func newFakeTokens() *fakeTokens {
	return &fakeTokens{now: time.Now(), expires: map[string]time.Time{}, users: map[string]int{}}
}

func (f *fakeTokens) StoreToken(_ context.Context, userID int, id, _ string, ttl time.Duration) error {
	f.expires[id], f.users[id] = f.now.Add(ttl), userID
	return nil
}
func (f *fakeTokens) GetToken(_ context.Context, _, id string) (int, error) {
	if exp, ok := f.expires[id]; !ok || !f.now.Before(exp) {
		return 0, shared_domain.ErrNotFound
	}
	return f.users[id], nil
}
func (f *fakeTokens) DeleteToken(_ context.Context, _, id string) error {
	delete(f.expires, id)
	return nil
}

// fakeJWT encodes the token id in the "token" string itself.
type fakeJWT struct{ shared_app.JwtService }

func (fakeJWT) GenerateToken(int, time.Duration) (string, error) { return "access", nil }
func (fakeJWT) GenerateRefreshToken(_ int, id string, _ time.Duration) (string, error) {
	return "refresh:" + id, nil
}

func (fakeJWT) ParseToken(token string) (*domain.CustomClaims, error) {
	return &domain.CustomClaims{UserID: 7, RegisteredClaims: jwt.RegisteredClaims{ID: token[len("refresh:"):]}}, nil
}

func newRefreshFixture() (*RefreshTokenUseCase, *fakeTokens, string) {
	tokens := newFakeTokens()
	_ = tokens.StoreToken(context.Background(), 7, "first", "refresh_token", RefreshTokenDuration)
	return NewRefreshTokenUseCase(fakeAuthRepo{}, tokens, fakeJWT{}), tokens, "refresh:first"
}

func TestRefreshRotatesAndSlidesTheWindow(t *testing.T) {
	uc, tokens, first := newRefreshFixture()
	tokens.now = tokens.now.Add(5 * 24 * time.Hour) // five idle days: still inside the 7

	session, err := uc.Execute(context.Background(), dtos.RefreshTokenInputDTO{RefreshToken: first})
	if err != nil {
		t.Fatalf("refresh after 5 days: %v", err)
	}
	if session.RefreshToken == first || session.RefreshToken == "" {
		t.Fatal("a new refresh token must be issued")
	}

	// The new token has a full lifetime again: six more days is fine, which a fixed 7-day
	// session from the first login would not allow.
	tokens.now = tokens.now.Add(6 * 24 * time.Hour)
	if _, err := uc.Execute(context.Background(), dtos.RefreshTokenInputDTO{RefreshToken: session.RefreshToken}); err != nil {
		t.Fatalf("the window should slide with activity: %v", err)
	}
}

func TestRefreshExpiresAfterAWeekOfInactivity(t *testing.T) {
	uc, tokens, first := newRefreshFixture()
	tokens.now = tokens.now.Add(7*24*time.Hour + time.Minute)

	if _, err := uc.Execute(context.Background(), dtos.RefreshTokenInputDTO{RefreshToken: first}); err != shared_domain.ErrInvalidToken {
		t.Fatalf("a week idle must end the session, got %v", err)
	}
}

func TestRefreshTokenSurvivesBrieflyForConcurrentTabs(t *testing.T) {
	uc, tokens, first := newRefreshFixture()

	if _, err := uc.Execute(context.Background(), dtos.RefreshTokenInputDTO{RefreshToken: first}); err != nil {
		t.Fatal(err)
	}
	// A second tab presents the same token a moment later.
	tokens.now = tokens.now.Add(5 * time.Second)
	if _, err := uc.Execute(context.Background(), dtos.RefreshTokenInputDTO{RefreshToken: first}); err != nil {
		t.Fatalf("a reuse inside the grace period must work: %v", err)
	}
	// But not once the grace period is over.
	tokens.now = tokens.now.Add(RefreshReuseGrace + time.Second)
	if _, err := uc.Execute(context.Background(), dtos.RefreshTokenInputDTO{RefreshToken: first}); err != shared_domain.ErrInvalidToken {
		t.Fatalf("a used token must die after the grace period, got %v", err)
	}
}
