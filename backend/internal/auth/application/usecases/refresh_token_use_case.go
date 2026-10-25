package usecases

import (
	"backend/internal/auth/application/dtos"
	"backend/internal/auth/domain"
	shared_app "backend/internal/shared/application"
	shared_domain "backend/internal/shared/domain"
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

// RefreshReuseGrace is how long a refresh token keeps working after it has been exchanged. Within
// it a replay gets the very same successor back, so two tabs (or a response lost on the network)
// never fork the session into two chains.
const RefreshReuseGrace = 30 * time.Second

// MaxSessionAge is the absolute lifetime of a login. Renewals slide the 7-day window but never
// past this limit, so a stolen session cannot be kept alive forever.
const MaxSessionAge = 30 * 24 * time.Hour

type RefreshTokenUseCase struct {
	authRepo   domain.AuthRepository
	tokenRepo  domain.TokenRepository
	jwtService shared_app.JwtService
	now        func() time.Time
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
		now:        time.Now,
	}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, input dtos.RefreshTokenInputDTO) (*dtos.AuthResponseDTO, error) {
	// 1. Parse the refresh token
	claims, err := uc.jwtService.ParseToken(input.RefreshToken)
	if err != nil || !isRefreshToken(claims) {
		return nil, shared_domain.ErrInvalidToken
	}

	// A password reset ends every session that started before it, even though those tokens are
	// otherwise still valid. Equal seconds count as revoked: tokens only carry whole seconds.
	revokedAt, err := uc.tokenRepo.SessionsRevokedAt(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if !revokedAt.IsZero() && claims.IssuedAt != nil && !claims.IssuedAt.Time.After(revokedAt) {
		return nil, shared_domain.ErrInvalidToken
	}

	// 2. Absolute session age. Legacy tokens (no session start) count from their own issue time.
	sessionStart := time.Unix(claims.SessionStart, 0)
	if claims.SessionStart == 0 {
		if claims.IssuedAt == nil {
			return nil, shared_domain.ErrInvalidToken
		}
		sessionStart = claims.IssuedAt.Time
	}
	now := uc.now()
	remaining := sessionStart.Add(MaxSessionAge).Sub(now)
	if remaining <= 0 {
		return nil, shared_domain.ErrInvalidToken
	}
	ttl := min(RefreshTokenDuration, remaining)

	// 3. Get user details to build the response (before rotating: a vanished account must not
	// leave a rotated chain behind).
	user, err := uc.authRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, shared_domain.ErrNotFound) {
			return nil, shared_domain.ErrInvalidToken
		}
		return nil, err
	}

	// 4. Exchange the token for its successor, atomically. Tokens issued before families existed
	// start theirs now, named after their own id.
	family := claims.Family
	legacy := family == ""
	if legacy {
		family = claims.ID
	}
	newRefreshTokenID := uuid.New().String()
	outcome, successorID, err := uc.tokenRepo.RotateRefreshToken(ctx, domain.RotateRequest{
		UserID:      claims.UserID,
		Family:      family,
		PresentedID: claims.ID,
		NewID:       newRefreshTokenID,
		Legacy:      legacy,
		Grace:       RefreshReuseGrace,
		TTL:         ttl,
	})
	if err != nil {
		return nil, err
	}
	switch outcome {
	case domain.RotateOK:
	case domain.RotateReplay:
		newRefreshTokenID = successorID // same successor as the first exchange
	case domain.RotateReuse:
		log.Printf("refresh token reuse detected for user %d: session family revoked", claims.UserID)
		return nil, shared_domain.ErrInvalidToken
	default:
		return nil, shared_domain.ErrInvalidToken
	}

	// 5. Generate the new access token and the successor's JWT
	newAccessToken, err := uc.jwtService.GenerateToken(user.ID, AccessTokenDuration)
	if err != nil {
		return nil, err
	}
	newRefreshToken, err := uc.jwtService.GenerateRefreshTokenInFamily(user.ID, newRefreshTokenID, family, sessionStart, ttl)
	if err != nil {
		return nil, err
	}

	// 6. Build and return the response
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

// isRefreshToken accepts tokens typed "refresh". Refresh tokens issued before tokens carried a
// type have no `typ` but do have a jti; that is still safe because step 2 of Execute requires the
// jti to be stored as a refresh token in Redis, which no access or verification token ever is
// (they carry no jti). It lets sessions open at the upgrade renew once instead of signing out;
// it can go once every such token has expired (7 days after the upgrade).
func isRefreshToken(claims *domain.CustomClaims) bool {
	return claims.Type == domain.TokenTypeRefresh || (claims.Type == "" && claims.ID != "")
}
