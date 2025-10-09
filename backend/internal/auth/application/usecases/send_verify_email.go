package usecases

import (
	"backend/internal/auth/application/dtos"
	"backend/internal/auth/domain"
	shared_app "backend/internal/shared/application"
	shared_domain "backend/internal/shared/domain"
	"context"
	"errors"
	"log"
)

type SendVerifyEmailUseCase struct {
	authRepo      domain.AuthRepository
	jwtService    shared_app.JwtService
	mailerService shared_app.MailerService
	store         shared_app.Store
}

func NewSendVerifyEmailUseCase(
	authRepo domain.AuthRepository,
	jwtS shared_app.JwtService,
	mailerService shared_app.MailerService,
	store shared_app.Store) *SendVerifyEmailUseCase {
	return &SendVerifyEmailUseCase{
		authRepo:      authRepo,
		jwtService:    jwtS,
		mailerService: mailerService,
		store:         store,
	}
}

func (uc *SendVerifyEmailUseCase) Execute(ctx context.Context, input dtos.SendVerifyEmailInputDTO) error {
	user, err := uc.authRepo.GetUserByEmailOrUsername(ctx, input.Email)
	if err != nil {
		if errors.Is(err, shared_domain.ErrNotFound) {
			return shared_domain.ErrInvalidCredentials // User not found
		}
		return err
	}

	message := shared_app.Message{
		To:      []string{user.Email},
		Subject: "Verify your email",
		Body:    "Please verify your email by clicking the link below.",
	}
	if err := uc.mailerService.Send(&message); err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	return nil
}
