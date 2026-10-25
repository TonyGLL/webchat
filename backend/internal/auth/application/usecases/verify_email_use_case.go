package usecases

import (
	"backend/internal/auth/application/dtos"
	"backend/internal/auth/domain"
	shared_app "backend/internal/shared/application"
	shared_domain "backend/internal/shared/domain"
	"context"
	"errors"
)

type VerifyEmailUseCase struct {
	authRepo      domain.AuthRepository
	tokenRepo     domain.TokenRepository
	jwtService    shared_app.JwtService
	mailerService shared_app.MailerService
}

func NewVerifyEmailUseCase(
	authRepo domain.AuthRepository,
	tokenRepo domain.TokenRepository,
	jwtService shared_app.JwtService,
	mailerService shared_app.MailerService) *VerifyEmailUseCase {
	return &VerifyEmailUseCase{
		authRepo:      authRepo,
		tokenRepo:     tokenRepo,
		jwtService:    jwtService,
		mailerService: mailerService,
	}
}

func (uc *VerifyEmailUseCase) Execute(ctx context.Context, input dtos.VerifyEmailInputDTO) error {
	userID, err := uc.tokenRepo.GetToken(ctx, "verify_email_token", input.Token)
	if err != nil {
		if errors.Is(err, shared_domain.ErrNotFound) {
			return shared_domain.ErrInvalidToken
		}
		return err
	}

	if err := uc.authRepo.VerifyEmail(ctx, userID); err != nil {
		return err
	}

	if err := uc.tokenRepo.DeleteToken(ctx, "verify_email_token", input.Token); err != nil {
		return err
	}

	return nil
}
