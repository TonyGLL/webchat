package usecases

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"backend/internal/auth/application/dtos"
	"backend/internal/auth/domain"
	shared_app "backend/internal/shared/application"
	shared_domain "backend/internal/shared/domain"
)

type resetAuthRepo struct {
	domain.AuthRepository
	users    map[string]*domain.User // by email
	password map[int]string
}

func (r *resetAuthRepo) GetUserByEmailOrUsername(_ context.Context, key string) (*domain.User, error) {
	for _, u := range r.users {
		if strings.EqualFold(u.Email, key) || u.Username == key {
			return u, nil
		}
	}
	return nil, shared_domain.ErrNotFound
}
func (r *resetAuthRepo) GetUserByID(_ context.Context, id int) (*domain.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, shared_domain.ErrNotFound
}
func (r *resetAuthRepo) UpdatePassword(_ context.Context, id int, hash string) error {
	r.password[id] = hash
	return nil
}

type fakeMailer struct{ sent []*shared_app.Message }

func (m *fakeMailer) Send(msg *shared_app.Message) error {
	m.sent = append(m.sent, msg)
	return nil
}

type plainHasher struct{}

func (plainHasher) HashPassword(p string) (string, error) { return "hash:" + p, nil }
func (plainHasher) CheckPasswordHash(p, h string) bool    { return h == "hash:"+p }

type resetFixture struct {
	request *RequestPasswordResetUseCase
	reset   *ResetPasswordUseCase
	repo    *resetAuthRepo
	tokens  *fakeTokens
	mailer  *fakeMailer
}

func newResetFixture() *resetFixture {
	repo := &resetAuthRepo{
		users:    map[string]*domain.User{"ana@example.com": {ID: 7, Name: "Ana", Username: "anita", Email: "ana@example.com"}},
		password: map[int]string{},
	}
	tokens := newFakeTokens()
	mailer := &fakeMailer{}
	return &resetFixture{
		request: NewRequestPasswordResetUseCase(repo, tokens, mailer, "https://chat.example.com/"),
		reset:   NewResetPasswordUseCase(repo, tokens, plainHasher{}),
		repo:    repo, tokens: tokens, mailer: mailer,
	}
}

// tokenFromMail extracts the secret from the emailed link.
func tokenFromMail(t *testing.T, m *shared_app.Message) string {
	t.Helper()
	const marker = "https://chat.example.com/reset-password/"
	i := strings.Index(m.Body, marker)
	if i < 0 {
		t.Fatalf("no reset link in the email: %q", m.Body)
	}
	rest := m.Body[i+len(marker):]
	if end := strings.IndexAny(rest, "\n \r"); end >= 0 {
		rest = rest[:end]
	}
	token, err := url.PathUnescape(rest)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestRequestResetStaysSilentAboutUnknownAddresses(t *testing.T) {
	f := newResetFixture()
	for _, email := range []string{"nobody@example.com", "anita"} { // unknown email; a username is not an email
		if err := f.request.Execute(context.Background(), email); err != nil {
			t.Fatalf("%s: must not error: %v", email, err)
		}
	}
	if len(f.mailer.sent) != 0 {
		t.Fatalf("no email may go out, got %d", len(f.mailer.sent))
	}
}

func TestRequestResetEmailsALinkAndStoresOnlyItsHash(t *testing.T) {
	f := newResetFixture()
	if err := f.request.Execute(context.Background(), "ANA@example.com"); err != nil {
		t.Fatal(err)
	}
	if len(f.mailer.sent) != 1 || f.mailer.sent[0].To[0] != "ana@example.com" {
		t.Fatalf("expected one email to ana, got %+v", f.mailer.sent)
	}
	token := tokenFromMail(t, f.mailer.sent[0])
	if _, ok := f.tokens.users[token]; ok {
		t.Fatal("the raw token must never be stored")
	}
	if _, ok := f.tokens.users[hashResetToken(token)]; !ok {
		t.Fatal("the token hash should be stored")
	}
}

func TestRequestResetIsThrottledPerAddress(t *testing.T) {
	f := newResetFixture()
	_ = f.request.Execute(context.Background(), "ana@example.com")
	_ = f.request.Execute(context.Background(), "ana@example.com")
	if len(f.mailer.sent) != 1 {
		t.Fatalf("a second request inside the cooldown must not send another email, got %d", len(f.mailer.sent))
	}
	f.tokens.now = f.tokens.now.Add(passwordResetCooldown + time.Second)
	_ = f.request.Execute(context.Background(), "ana@example.com")
	if len(f.mailer.sent) != 2 {
		t.Fatalf("after the cooldown a new link may be sent, got %d", len(f.mailer.sent))
	}
}

func TestResetPasswordHappyPathIsSingleUseAndEndsSessions(t *testing.T) {
	f := newResetFixture()
	_ = f.request.Execute(context.Background(), "ana@example.com")
	token := tokenFromMail(t, f.mailer.sent[0])

	if err := f.reset.Execute(context.Background(), token, "a-brand-new-secret"); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if f.repo.password[7] != "hash:a-brand-new-secret" {
		t.Fatal("the new password hash must be stored")
	}
	if f.tokens.revokedAt.IsZero() {
		t.Fatal("open sessions must be revoked")
	}
	if err := f.reset.Execute(context.Background(), token, "another-secret-1"); !errors.Is(err, shared_domain.ErrInvalidToken) {
		t.Fatalf("the link must work only once, got %v", err)
	}
}

func TestResetPasswordRejectsBadTokensAndKeepsLinkOnWeakPassword(t *testing.T) {
	f := newResetFixture()
	if err := f.reset.Execute(context.Background(), "made-up", "a-brand-new-secret"); !errors.Is(err, shared_domain.ErrInvalidToken) {
		t.Fatalf("unknown token: got %v", err)
	}

	_ = f.request.Execute(context.Background(), "ana@example.com")
	token := tokenFromMail(t, f.mailer.sent[0])
	if err := f.reset.Execute(context.Background(), token, "anita-2024"); !errors.Is(err, domain.ErrPasswordPersonal) {
		t.Fatalf("a password containing the username must be refused, got %v", err)
	}
	if err := f.reset.Execute(context.Background(), token, "a-brand-new-secret"); err != nil {
		t.Fatalf("after a refused password the same link must still work: %v", err)
	}
}

func TestResetLinkExpires(t *testing.T) {
	f := newResetFixture()
	_ = f.request.Execute(context.Background(), "ana@example.com")
	token := tokenFromMail(t, f.mailer.sent[0])
	f.tokens.now = f.tokens.now.Add(PasswordResetTTL + time.Second)
	if err := f.reset.Execute(context.Background(), token, "a-brand-new-secret"); !errors.Is(err, shared_domain.ErrInvalidToken) {
		t.Fatalf("an expired link must be refused, got %v", err)
	}
}

func TestPasswordResetEndsExistingRefreshSessions(t *testing.T) {
	f := newResetFixture()
	refresh := NewRefreshTokenUseCase(fakeAuthRepo{}, f.tokens, fakeJWT{})
	_ = f.tokens.StoreToken(context.Background(), 7, "old", "refresh_token", RefreshTokenDuration)

	// Issued "a minute ago": the fake JWT stamps iat from the claims below.
	f.tokens.now = f.tokens.now.Add(time.Minute)
	_ = f.request.Execute(context.Background(), "ana@example.com")
	token := tokenFromMail(t, f.mailer.sent[0])
	if err := f.reset.Execute(context.Background(), token, "a-brand-new-secret"); err != nil {
		t.Fatal(err)
	}
	if _, err := refresh.Execute(context.Background(), dtos.RefreshTokenInputDTO{RefreshToken: "refresh:old"}); !errors.Is(err, shared_domain.ErrInvalidToken) {
		t.Fatalf("a session opened before the reset must be refused, got %v", err)
	}
}
