package usecases

import (
	"context"
	"log"
	"time"

	auth_usecases "backend/internal/auth/application/usecases"
	auth_domain "backend/internal/auth/domain"
	shared_domain "backend/internal/shared/domain"
	"backend/internal/users/domain"
)

// SessionRevoker ends every open session of a user (satisfied by the auth token repository).
type SessionRevoker interface {
	RevokeUserSessions(ctx context.Context, userID int, ttl time.Duration) error
}

// DeactivateUserUseCase disables an account after the user confirms their password, and ends
// every session and live connection they have open.
type DeactivateUserUseCase struct {
	usersRepo domain.UsersRepository
	passwords auth_domain.PasswordService
	sessions  SessionRevoker
	realtime  Realtime
}

func NewDeactivateUserUseCase(
	usersRepo domain.UsersRepository,
	passwords auth_domain.PasswordService,
	sessions SessionRevoker,
	realtime Realtime) *DeactivateUserUseCase {
	return &DeactivateUserUseCase{
		usersRepo: usersRepo,
		passwords: passwords,
		sessions:  sessions,
		realtime:  realtime,
	}
}

// Execute deactivates the account. A wrong password returns ErrInvalidCredentials.
func (uc *DeactivateUserUseCase) Execute(ctx context.Context, userID int, password string) error {
	hash, err := uc.usersRepo.PasswordHash(ctx, userID)
	if err != nil {
		return err
	}
	if !uc.passwords.CheckPasswordHash(password, hash) {
		return shared_domain.ErrInvalidCredentials
	}

	if err := uc.usersRepo.DeactivateUser(ctx, userID); err != nil {
		return err
	}

	// The account is already off; the revocation also kills access tokens issued before this
	// moment and refresh tokens. A Redis failure is logged, not surfaced: the deleted flag still
	// blocks the account (within the guard's cache window).
	if err := uc.sessions.RevokeUserSessions(ctx, userID, auth_usecases.RefreshTokenDuration); err != nil {
		log.Printf("failed to revoke sessions of deactivated user %d: %v", userID, err)
	}
	if uc.realtime != nil {
		uc.realtime.DisconnectUser(userID)
	}
	return nil
}
