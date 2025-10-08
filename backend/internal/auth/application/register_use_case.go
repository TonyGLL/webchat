package application

import (
	"context"
	"time"

	"backend/internal/auth/domain"
	shared_app "backend/internal/shared/application"
)

type RegisterUseCase struct {
	store           shared_app.Store
	passwordService domain.PasswordService
	jwtService      shared_app.JwtService
}

func NewRegisterUseCase(
	store shared_app.Store,
	ps domain.PasswordService,
	jwtS shared_app.JwtService) *RegisterUseCase {
	return &RegisterUseCase{
		store:           store,
		passwordService: ps,
		jwtService:      jwtS,
	}
}

// Execute handles user registration, ensuring all steps are performed atomically.
// It returns the newly created user or an error.
func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInputDTO) (*AuthResponseDTO, error) {
	hashedPassword, err := uc.passwordService.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:    input.Email,
		Name:     input.Name,
		LastName: input.LastName,
		Username: input.Username,
		Phone:    input.Phone,
	}

	var createdUser *domain.User

	// Execute the registration within a single transaction
	err = uc.store.ExecTx(ctx, func(store shared_app.Store) error {
		authRepo := store.AuthRepository()

		// 1. Register the user
		newUser, err := authRepo.Register(ctx, user)
		if err != nil {
			return err
		}

		// 2. Create the password entry
		if err := authRepo.CreatePassword(ctx, newUser.ID, hashedPassword); err != nil {
			return err
		}

		// 3. Set initial last access time
		if err := authRepo.SetLastAccess(ctx, newUser.ID); err != nil {
			// Note: Depending on business rules, this might not be a fatal error.
			// For now, it's included in the transaction.
			return err
		}

		createdUser = newUser
		return nil
	})

	if err != nil {
		return nil, err
	}

	token, err := uc.jwtService.GenerateToken(createdUser.ID, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	response := uc.buildRegisterResponse(createdUser, token)

	return &response, nil
}

func (uc *RegisterUseCase) buildRegisterResponse(user *domain.User, token string) AuthResponseDTO {
	userResponse := UserResponseDTO{
		ID:        user.ID,
		Name:      user.Name,
		LastName:  user.LastName,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		AvatarURL: user.AvatarUrl,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	return AuthResponseDTO{
		User:  userResponse,
		Token: token,
	}
}
