package usecases

import (
	"backend/internal/auth/domain"
	shared_app "backend/internal/shared/application"
	"context"
)

type SendVerifyEmailUseCase struct {
	authRepo   domain.AuthRepository
	jwtService shared_app.JwtService
	store      shared_app.Store
}

func NewSendVerifyEmailUseCase(
	authRepo domain.AuthRepository,
	jwtS shared_app.JwtService) *SendVerifyEmailUseCase {
	return &SendVerifyEmailUseCase{
		authRepo:   authRepo,
		jwtService: jwtS,
	}
}

func (uc *SendVerifyEmailUseCase) Execute(ctx context.Context, userID int) error {
	return nil
}
