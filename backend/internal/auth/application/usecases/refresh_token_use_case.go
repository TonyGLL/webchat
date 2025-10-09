package usecases

import (
	"backend/internal/auth/application/dtos"
	"backend/internal/auth/domain"
	shared_app "backend/internal/shared/application"
	shared_domain "backend/internal/shared/domain"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type RefreshTokenUseCase struct {
	authRepo   domain.AuthRepository
	tokenRepo  domain.TokenRepository
	jwtService shared_app.JwtService
}

func NewRefreshTokenUseCase(
	authRepo domain.AuthRepository,
	tokenRepo domain.TokenRepository,
	jwtService shared_app.JwtService,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		authRepo:   authRepo,
		tokenRepo:  tokenRepo,
		jwtService: jwtService,
	}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, input dtos.RefreshTokenInputDTO) (*dtos.AuthResponseDTO, error) {
	// 1. Parse the refresh token
	claims, err := uc.jwtService.ParseToken(input.RefreshToken)
	if err != nil {
		return nil, shared_domain.ErrInvalidToken
	}

	// 2. Verify the token exists in our store (Redis)
	storedUserID, err := uc.tokenRepo.GetToken(ctx, "refresh_token", claims.ID)
	if err != nil {
		if errors.Is(err, shared_domain.ErrNotFound) {
			return nil, shared_domain.ErrInvalidToken // Token not found or already used
		}
		return nil, err
	}

	// 3. Check if the user ID from the token matches the one in the store
	if claims.UserID != storedUserID {
		return nil, shared_domain.ErrInvalidToken
	}

	// 4. Delete the old refresh token to prevent reuse
	if err := uc.tokenRepo.DeleteToken(ctx, "refresh_token", claims.ID); err != nil {
		// Log the error but continue, as the user should still get a new token
		fmt.Printf("Warning: failed to delete old refresh token: %v\n", err)
	}

	// 5. Get user details to build the response
	user, err := uc.authRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, shared_domain.ErrNotFound) {
			// This case is unlikely if the token was valid, but handle it for safety.
			return nil, shared_domain.ErrInvalidToken
		}
		return nil, err
	}

	// 6. Generate a new pair of tokens
	newAccessToken, err := uc.jwtService.GenerateToken(user.ID, AccessTokenDuration)
	if err != nil {
		return nil, err
	}

	newRefreshTokenID := uuid.New().String()
	newRefreshToken, err := uc.jwtService.GenerateRefreshToken(user.ID, newRefreshTokenID, RefreshTokenDuration)
	if err != nil {
		return nil, err
	}

	// 7. Store the new refresh token
	if err := uc.tokenRepo.StoreToken(ctx, user.ID, newRefreshTokenID, "refresh_token", RefreshTokenDuration); err != nil {
		return nil, err
	}

	// 8. Build and return the response
	userResponse := dtos.UserResponseDTO{
		ID:       user.ID,
		Name:     user.Name,
		LastName: user.LastName,
		Username: user.Username,
		Email:    user.Email,
	}

	return &dtos.AuthResponseDTO{
		User:         userResponse,
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
