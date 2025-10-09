package usecases

import (
	"context"

	"backend/internal/auth/application/dtos"
	"backend/internal/auth/domain"
	shared_app "backend/internal/shared/application"
)

type RegisterUseCase struct {
	authRepo        domain.AuthRepository
	passwordService domain.PasswordService
	jwtService      shared_app.JwtService
	store           shared_app.Store
}

func NewRegisterUseCase(
	authRepo domain.AuthRepository,
	ps domain.PasswordService,
	jwtS shared_app.JwtService,
	store shared_app.Store) *RegisterUseCase {
	return &RegisterUseCase{
		authRepo:        authRepo,
		passwordService: ps,
		jwtService:      jwtS,
		store:           store,
	}
}

// Execute handles user registration, ensuring all steps are performed atomically.
// It returns the newly created user or an error.
func (uc *RegisterUseCase) Execute(ctx context.Context, input dtos.RegisterInputDTO) (*dtos.UserResponseDTO, error) {
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
		// 1. Register the user
		newUser, err := uc.authRepo.Register(ctx, user)
		if err != nil {
			return err
		}

		// 2. Create the password entry
		if err := uc.authRepo.CreatePassword(ctx, newUser.ID, hashedPassword); err != nil {
			return err
		}

		// 3. Set initial last access time
		if err := uc.authRepo.SetLastAccess(ctx, newUser.ID); err != nil {
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

	response := dtos.UserResponseDTO{
		ID:    createdUser.ID,
		Email: createdUser.Email,
	}

	return &response, nil
}
