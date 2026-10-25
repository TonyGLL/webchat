package usecases

import (
	"context"

	"backend/internal/auth/domain"
	shared_app "backend/internal/shared/application"
)

// LogoutUseCase revokes a refresh token so it can no longer be exchanged for new sessions.
type LogoutUseCase struct {
	tokenRepo  domain.TokenRepository
	jwtService shared_app.JwtService
	realtime   Realtime
}

// NewLogoutUseCase builds the use case; realtime may be nil.
func NewLogoutUseCase(tokenRepo domain.TokenRepository, jwtService shared_app.JwtService, realtime Realtime) *LogoutUseCase {
	return &LogoutUseCase{tokenRepo: tokenRepo, jwtService: jwtService, realtime: realtime}
}

// Execute is best-effort and idempotent: an unknown or expired token is already "logged out".
func (uc *LogoutUseCase) Execute(ctx context.Context, refreshToken string) error {
	claims, err := uc.jwtService.ParseToken(refreshToken)
	if err != nil || !isRefreshToken(claims) || claims.ID == "" {
		return nil
	}
	err = uc.tokenRepo.DeleteToken(ctx, "refresh_token", claims.ID)
	// Sockets are not tied to a session, so the user's open sockets are cut (other tabs
	// reconnect on their own with a fresh ticket) instead of listening on after the logout.
	if uc.realtime != nil {
		uc.realtime.DisconnectUser(claims.UserID)
	}
	return err
}
