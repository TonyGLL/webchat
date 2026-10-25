package usecases

import (
	"backend/internal/auth/domain"
	shared_app "backend/internal/shared/application"
	shared_domain "backend/internal/shared/domain"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	passwordResetKey = "password_reset"
	// PasswordResetTTL is how long an emailed reset link works.
	PasswordResetTTL = 30 * time.Minute
	// passwordResetCooldown stops one address from being mailed over and over.
	passwordResetCooldown    = time.Minute
	passwordResetCooldownKey = "password_reset_cooldown"
)

// hashResetToken is what gets stored: if Redis leaked, the stored values could not be used
// as links.
func hashResetToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// RequestPasswordResetUseCase emails a single-use link. It never reveals whether the address
// has an account: every outcome looks the same to the caller.
type RequestPasswordResetUseCase struct {
	authRepo      domain.AuthRepository
	tokenRepo     domain.TokenRepository
	mailerService shared_app.MailerService
	frontendURL   string
	// dispatch runs the mail delivery off the request, so the response time does not reveal
	// whether the address has an account. Tests replace it with a synchronous call.
	dispatch func(func())
}

// mailSendTimeout bounds one background delivery attempt.
const mailSendTimeout = 30 * time.Second

func NewRequestPasswordResetUseCase(authRepo domain.AuthRepository, tokenRepo domain.TokenRepository, mailer shared_app.MailerService, frontendURL string) *RequestPasswordResetUseCase {
	return &RequestPasswordResetUseCase{authRepo: authRepo, tokenRepo: tokenRepo, mailerService: mailer, frontendURL: frontendURL,
		dispatch: func(fn func()) { go fn() }}
}

func (uc *RequestPasswordResetUseCase) Execute(ctx context.Context, email string) error {
	email = domain.NormalizeEmail(email)
	user, err := uc.authRepo.GetUserByEmailOrUsername(ctx, email)
	if err != nil {
		if errors.Is(err, shared_domain.ErrNotFound) {
			return nil
		}
		return err
	}
	// The lookup also matches usernames: only an exact email match may receive a reset.
	if user.Deleted || !strings.EqualFold(user.Email, email) {
		return nil
	}

	cooldownID := hashResetToken(strings.ToLower(user.Email))
	if _, err := uc.tokenRepo.GetToken(ctx, passwordResetCooldownKey, cooldownID); err == nil {
		return nil // a link was sent a moment ago
	} else if !errors.Is(err, shared_domain.ErrNotFound) {
		return err
	}
	if err := uc.tokenRepo.StoreToken(ctx, user.ID, cooldownID, passwordResetCooldownKey, passwordResetCooldown); err != nil {
		return err
	}

	token, err := newResetToken()
	if err != nil {
		return err
	}
	if err := uc.tokenRepo.StoreExclusiveToken(ctx, user.ID, hashResetToken(token), passwordResetKey, PasswordResetTTL); err != nil {
		return err
	}

	link := buildResetURL(uc.frontendURL, token)
	minutes := int(PasswordResetTTL.Minutes())
	message := shared_app.Message{
		To:      []string{user.Email},
		Subject: "Restablecer contraseña / Reset your password",
		Body: "Hola,\n\nPara elegir una contraseña nueva abre este enlace (válido " + strconv.Itoa(minutes) + " minutos, un solo uso):\n" + link +
			"\n\nSi no lo pediste, ignora este mensaje: tu contraseña no cambia.\n\n---\n\n" +
			"Hi,\n\nTo choose a new password open this link (valid for " + strconv.Itoa(minutes) + " minutes, single use):\n" + link +
			"\n\nIf you did not ask for this, ignore this message: your password stays the same.",
	}
	uc.dispatch(func() {
		// The mailer takes no context: give up waiting after a while so the goroutine ends.
		sent := make(chan error, 1)
		go func() { sent <- uc.mailerService.Send(&message) }()
		select {
		case err := <-sent:
			// Do not tell the caller: a failure here must look like any other outcome.
			if err != nil {
				log.Printf("failed to send password reset email: %v", err)
			}
		case <-time.After(mailSendTimeout):
			log.Printf("password reset email timed out after %s", mailSendTimeout)
		}
	})
	return nil
}

func buildResetURL(frontendURL, token string) string {
	return strings.TrimRight(frontendURL, "/") + "/reset-password/" + url.PathEscape(token)
}

// ResetPasswordUseCase sets the new password with a valid link, ends every open session and
// burns the link.
type ResetPasswordUseCase struct {
	authRepo        domain.AuthRepository
	tokenRepo       domain.TokenRepository
	passwordService domain.PasswordService
	realtime        Realtime
}

// Realtime lets a password reset cut the user's open websocket connections.
type Realtime interface {
	DisconnectUser(userID int)
}

// NewResetPasswordUseCase builds the use case; realtime may be nil.
func NewResetPasswordUseCase(authRepo domain.AuthRepository, tokenRepo domain.TokenRepository, ps domain.PasswordService, realtime Realtime) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{authRepo: authRepo, tokenRepo: tokenRepo, passwordService: ps, realtime: realtime}
}

func (uc *ResetPasswordUseCase) Execute(ctx context.Context, token, newPassword string) error {
	tokenID := hashResetToken(token)
	// Peek first: a refused password must leave the link usable.
	userID, err := uc.tokenRepo.GetToken(ctx, passwordResetKey, tokenID)
	if err != nil {
		if errors.Is(err, shared_domain.ErrNotFound) {
			return shared_domain.ErrInvalidToken
		}
		return err
	}

	user, err := uc.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, shared_domain.ErrNotFound) {
			return shared_domain.ErrInvalidToken
		}
		return err
	}
	if err := domain.ValidatePassword(newPassword, user.Username, user.Email); err != nil {
		return err // the link stays valid so the person can try another password
	}

	hashed, err := uc.passwordService.HashPassword(newPassword)
	if err != nil {
		return err
	}
	// Burn the link atomically before changing anything: of two concurrent redemptions only one
	// gets past this point. (If the update below then fails, the person asks for a new link.)
	if _, err := uc.tokenRepo.ConsumeToken(ctx, passwordResetKey, tokenID); err != nil {
		if errors.Is(err, shared_domain.ErrNotFound) {
			return shared_domain.ErrInvalidToken
		}
		return err
	}
	if err := uc.authRepo.UpdatePassword(ctx, userID, hashed); err != nil {
		return err
	}
	// Whoever had the old password (or a stolen session) is signed out everywhere.
	if err := uc.tokenRepo.RevokeUserSessions(ctx, userID, RefreshTokenDuration); err != nil {
		return err
	}
	// ...and so are the sockets they already opened: the access token check only runs per request.
	if uc.realtime != nil {
		uc.realtime.DisconnectUser(userID)
	}
	return nil
}
