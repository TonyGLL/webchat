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
}

func NewLogoutUseCase(tokenRepo domain.TokenRepository, jwtService shared_app.JwtService) *LogoutUseCase {
	return &LogoutUseCase{tokenRepo: tokenRepo, jwtService: jwtService}
}

// Execute is best-effort and idempotent: an unknown or expired token is already "logged out".
func (uc *LogoutUseCase) Execute(ctx context.Context, refreshToken string) error {
	claims, err := uc.jwtService.ParseToken(refreshToken)
	if err != nil || !isRefreshToken(claims) || claims.ID == "" {
		return nil
	}
	return uc.tokenRepo.DeleteToken(ctx, "refresh_token", claims.ID)
}
