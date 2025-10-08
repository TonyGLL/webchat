package usecases

import (
	"backend/application"
	"backend/application/dtos"
	"backend/domain"
	"context"
)

type RegisterUseCase struct {
	store           application.Store
	passwordService domain.PasswordService
}

func NewRegisterUseCase(store application.Store, ps domain.PasswordService) *RegisterUseCase {
	return &RegisterUseCase{
		store:           store,
		passwordService: ps,
	}
}

// Execute handles user registration, ensuring all steps are performed atomically.
// It returns the newly created user or an error.
func (uc *RegisterUseCase) Execute(ctx context.Context, input dtos.RegisterInputDTO) (*domain.User, error) {
	hashedPassword, err := uc.passwordService.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:    input.Email,
		Name:     input.Name,
		LastName: input.LastName,
		Username: input.Username, // Corrected from UserName
		Phone:    input.Phone,
	}

	var createdUser *domain.User

	// Execute the registration within a single transaction
	err = uc.store.ExecTx(ctx, func(store application.Store) error {
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

	return createdUser, nil
}