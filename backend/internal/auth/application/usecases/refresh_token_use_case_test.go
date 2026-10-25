package usecases

import (
	"context"
	"strconv"
	"strings"
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
	now       time.Time
	expires   map[string]time.Time
	users     map[string]int
	revokedAt time.Time
	// rotation state: family -> current token id, rotated token id -> successor id.
	family  map[string]string
	rotated map[string]string
	// exclusive holds the per-user pointer used by StoreExclusiveToken.
	exclusive map[int]string
}

func newFakeTokens() *fakeTokens {
	return &fakeTokens{now: time.Now(), expires: map[string]time.Time{}, users: map[string]int{},
		family: map[string]string{}, rotated: map[string]string{}, exclusive: map[int]string{}}
}

func (f *fakeTokens) ConsumeToken(ctx context.Context, key, id string) (int, error) {
	uid, err := f.GetToken(ctx, key, id)
	if err != nil {
		return 0, err
	}
	delete(f.expires, id)
	return uid, nil
}

func (f *fakeTokens) StoreExclusiveToken(ctx context.Context, userID int, id, key string, ttl time.Duration) error {
	if old, ok := f.exclusive[userID]; ok {
		delete(f.expires, old)
	}
	f.exclusive[userID] = id
	return f.StoreToken(ctx, userID, id, key, ttl)
}

// RotateRefreshToken mirrors the Redis script (see redis_token_repository.go).
func (f *fakeTokens) RotateRefreshToken(ctx context.Context, r domain.RotateRequest) (domain.RotateOutcome, string, error) {
	current, hasFamily := f.family[r.Family]
	if !r.Legacy && !hasFamily {
		return domain.RotateInvalid, "", nil
	}
	uid, err := f.GetToken(ctx, "refresh_token", r.PresentedID)
	if err != nil {
		if _, was := f.rotated[r.PresentedID]; was {
			delete(f.family, r.Family)
			return domain.RotateReuse, "", nil
		}
		return domain.RotateInvalid, "", nil
	}
	if uid != r.UserID {
		return domain.RotateInvalid, "", nil
	}
	if succ, ok := f.rotated[r.PresentedID]; ok {
		return domain.RotateReplay, succ, nil
	}
	if hasFamily && current != r.PresentedID {
		return domain.RotateInvalid, "", nil
	}
	f.rotated[r.PresentedID] = r.NewID
	f.expires[r.PresentedID] = f.now.Add(r.Grace)
	f.expires[r.NewID], f.users[r.NewID] = f.now.Add(r.TTL), r.UserID
	f.family[r.Family] = r.NewID
	return domain.RotateOK, "", nil
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
func (f *fakeTokens) RevokeUserSessions(context.Context, int, time.Duration) error {
	f.revokedAt = f.now
	return nil
}
func (f *fakeTokens) SessionsRevokedAt(context.Context, int) (time.Time, error) {
	return f.revokedAt, nil
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
func (fakeJWT) GenerateRefreshTokenInFamily(_ int, id, family string, start time.Time, _ time.Duration) (string, error) {
	return "refresh:" + id + "@" + family + "@" + strconv.FormatInt(start.Unix(), 10), nil
}

// tokenIssuedAt is the fake issue time of every refresh token: before any clock the tests use.
var tokenIssuedAt = time.Now().Add(-time.Hour)

// ParseToken stamps every token as issued an hour before the tests start, i.e. before any
// revocation moment they create.
func (fakeJWT) ParseToken(token string) (*domain.CustomClaims, error) {
	if strings.HasPrefix(token, "access") {
		return &domain.CustomClaims{UserID: 7, Type: domain.TokenTypeAccess}, nil
	}
	if strings.HasPrefix(token, "verify") {
		return &domain.CustomClaims{UserID: 7, Type: domain.TokenTypeVerify}, nil
	}
	if strings.HasPrefix(token, "legacy-nojti") {
		return &domain.CustomClaims{UserID: 7}, nil
	}
	if strings.HasPrefix(token, "legacy:") {
		return &domain.CustomClaims{UserID: 7, RegisteredClaims: jwt.RegisteredClaims{
			ID:       token[len("legacy:"):],
			IssuedAt: jwt.NewNumericDate(tokenIssuedAt),
		}}, nil
	}
	claims := &domain.CustomClaims{UserID: 7, Type: domain.TokenTypeRefresh, RegisteredClaims: jwt.RegisteredClaims{
		ID:       token[len("refresh:"):],
		IssuedAt: jwt.NewNumericDate(tokenIssuedAt),
	}}
	if parts := strings.Split(claims.ID, "@"); len(parts) == 3 {
		claims.ID, claims.Family = parts[0], parts[1]
		claims.SessionStart, _ = strconv.ParseInt(parts[2], 10, 64)
	}
	return claims, nil
}

func newRefreshFixture() (*RefreshTokenUseCase, *fakeTokens, string) {
	tokens := newFakeTokens()
	_ = tokens.StoreToken(context.Background(), 7, "first", "refresh_token", RefreshTokenDuration)
	uc := NewRefreshTokenUseCase(fakeAuthRepo{}, tokens, fakeJWT{})
	uc.now = func() time.Time { return tokens.now }
	return uc, tokens, "refresh:first"
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

func TestRefreshRejectsTokensThatAreNotRefreshTokens(t *testing.T) {
	uc, _, _ := newRefreshFixture()
	if _, err := uc.Execute(context.Background(), dtos.RefreshTokenInputDTO{RefreshToken: "access"}); err != shared_domain.ErrInvalidToken {
		t.Fatalf("an access token must not work as a refresh token, got %v", err)
	}
}

// Sessions that were open before tokens carried a type must renew once, not sign out.
func TestRefreshAcceptsUntypedRefreshTokensStoredInRedis(t *testing.T) {
	uc, _, _ := newRefreshFixture()
	if _, err := uc.Execute(context.Background(), dtos.RefreshTokenInputDTO{RefreshToken: "legacy:first"}); err != nil {
		t.Fatalf("an untyped refresh token with a stored jti should renew, got %v", err)
	}
}

func TestRefreshStillRejectsUntypedAndVerifyTokensWithoutAStoredJti(t *testing.T) {
	uc, _, _ := newRefreshFixture()
	for _, token := range []string{"legacy-nojti", "verify", "legacy:not-stored"} {
		if _, err := uc.Execute(context.Background(), dtos.RefreshTokenInputDTO{RefreshToken: token}); err != shared_domain.ErrInvalidToken {
			t.Fatalf("%s must be rejected, got %v", token, err)
		}
	}
}

func refresh(uc *RefreshTokenUseCase, token string) (*dtos.AuthResponseDTO, error) {
	return uc.Execute(context.Background(), dtos.RefreshTokenInputDTO{RefreshToken: token})
}

// Two tabs replaying the same token inside the grace must land on one chain, not fork it.
func TestRefreshReplayInsideGraceReturnsTheSameSuccessor(t *testing.T) {
	uc, tokens, first := newRefreshFixture()
	a, err := refresh(uc, first)
	if err != nil {
		t.Fatal(err)
	}
	tokens.now = tokens.now.Add(20 * time.Second)
	b, err := refresh(uc, first)
	if err != nil {
		t.Fatalf("replay inside the grace: %v", err)
	}
	idOf := func(tok string) string { return strings.SplitN(strings.TrimPrefix(tok, "refresh:"), "@", 2)[0] }
	if idOf(a.RefreshToken) != idOf(b.RefreshToken) {
		t.Fatalf("a replay must hand back the same successor: %q vs %q", a.RefreshToken, b.RefreshToken)
	}
	// Replaying every 20 s must not keep the old token alive: the grace is not re-armed.
	tokens.now = tokens.now.Add(20 * time.Second)
	if _, err := refresh(uc, first); err != shared_domain.ErrInvalidToken {
		t.Fatalf("the grace must not slide with replays, got %v", err)
	}
}

// A rotated token coming back after the grace means a copy leaked: the family dies, including
// the token the legitimate client holds now.
func TestRefreshReuseAfterGraceRevokesTheFamily(t *testing.T) {
	uc, tokens, first := newRefreshFixture()
	victim, err := refresh(uc, first)
	if err != nil {
		t.Fatal(err)
	}
	tokens.now = tokens.now.Add(RefreshReuseGrace + time.Second)
	if _, err := refresh(uc, first); err != shared_domain.ErrInvalidToken {
		t.Fatalf("reuse after the grace must be refused, got %v", err)
	}
	if _, err := refresh(uc, victim.RefreshToken); err != shared_domain.ErrInvalidToken {
		t.Fatalf("the whole family must be revoked, got %v", err)
	}
}

func TestRefreshChainKeepsWorkingAcrossRotations(t *testing.T) {
	uc, tokens, tok := newRefreshFixture()
	for i := 0; i < 5; i++ {
		tokens.now = tokens.now.Add(time.Hour)
		session, err := refresh(uc, tok)
		if err != nil {
			t.Fatalf("rotation %d: %v", i, err)
		}
		tok = session.RefreshToken
	}
}

func TestRefreshSessionHasAnAbsoluteMaximumAge(t *testing.T) {
	uc, tokens, tok := newRefreshFixture()
	var lastErr error
	days := 0
	// Active every 5 days: the 7-day window would slide forever without the absolute limit.
	for ; days < 60; days += 5 {
		tokens.now = tokens.now.Add(5 * 24 * time.Hour)
		session, err := refresh(uc, tok)
		if err != nil {
			lastErr = err
			break
		}
		tok = session.RefreshToken
	}
	if lastErr != shared_domain.ErrInvalidToken {
		t.Fatalf("the session must end at the maximum age, got %v", lastErr)
	}
	if days < 25 || days > 35 {
		t.Fatalf("the session ended after %d days, want about %d", days, int(MaxSessionAge.Hours()/24))
	}
}

// Revocation and a token minted in the same second: the token must lose.
func TestRefreshRejectsTokensIssuedInTheSecondOfTheRevocation(t *testing.T) {
	uc, tokens, first := newRefreshFixture()
	tokens.revokedAt = tokenIssuedAt
	if _, err := refresh(uc, first); err != shared_domain.ErrInvalidToken {
		t.Fatalf("issued at the revocation instant must be refused, got %v", err)
	}
}

func TestLogoutCutsTheUsersSockets(t *testing.T) {
	_, tokens, first := newRefreshFixture()
	rt := &fakeRealtime{}
	if err := NewLogoutUseCase(tokens, fakeJWT{}, rt).Execute(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	if len(rt.disconnected) != 1 || rt.disconnected[0] != 7 {
		t.Fatalf("logout must cut the user's sockets, got %v", rt.disconnected)
	}
	if err := NewLogoutUseCase(tokens, fakeJWT{}, rt).Execute(context.Background(), "garbage-not-parsed"); err != nil {
		t.Fatal(err)
	}
}
