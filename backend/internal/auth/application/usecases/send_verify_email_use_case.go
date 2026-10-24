package usecases

import (
	"backend/internal/auth/application/dtos"
	"backend/internal/auth/domain"
	shared_app "backend/internal/shared/application"
	shared_domain "backend/internal/shared/domain"
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"
)

type SendVerifyEmailUseCase struct {
	authRepo      domain.AuthRepository
	tokenRepo     domain.TokenRepository
	jwtService    shared_app.JwtService
	mailerService shared_app.MailerService
	store         shared_app.Store
	frontendURL   string
}

func NewSendVerifyEmailUseCase(
	authRepo domain.AuthRepository,
	tokenRepo domain.TokenRepository,
	jwtS shared_app.JwtService,
	mailerService shared_app.MailerService,
	store shared_app.Store,
	frontendURL string) *SendVerifyEmailUseCase {
	return &SendVerifyEmailUseCase{
		authRepo:      authRepo,
		tokenRepo:     tokenRepo,
		jwtService:    jwtS,
		mailerService: mailerService,
		store:         store,
		frontendURL:   frontendURL,
	}
}

func (uc *SendVerifyEmailUseCase) Execute(ctx context.Context, input dtos.SendVerifyEmailInputDTO) error {
	user, err := uc.authRepo.GetUserVerifyEmail(ctx, input.Email, input.ID)
	if err != nil {
		if errors.Is(err, shared_domain.ErrNotFound) {
			return shared_domain.ErrNotFound // User not found
		}
		return err
	}

	tokenDuration := 10 * time.Minute
	// Generate a verification token
	claims := map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"typ":     domain.TokenTypeVerify,
	}
	token, err := uc.jwtService.GenerateTokenFromClaims(claims, tokenDuration) // 10 minutes
	if err != nil {
		return err
	}

	// Store the token in the store with an expiration time
	if err := uc.tokenRepo.StoreToken(ctx, user.ID, token, "verify_email_token", tokenDuration); err != nil {
		return err
	}

	verifyURL := buildVerifyURL(uc.frontendURL, token)
	message := shared_app.Message{
		To:      []string{user.Email},
		Subject: "Verify your email",
		Body:    "Please verify your email by clicking the following link: " + verifyURL,
	}
	if err := uc.mailerService.Send(&message); err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	return nil
}

// buildVerifyURL returns the frontend page that confirms the email for the given token.
func buildVerifyURL(frontendURL, token string) string {
	return fmt.Sprintf("%s/verify/%s", strings.TrimRight(frontendURL, "/"), url.PathEscape(token))
}
