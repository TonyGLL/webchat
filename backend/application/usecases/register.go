package usecases

import (
	"backend/application/dtos"
	"backend/application/repositories"
	"backend/domain"
	"context"

	"golang.org/x/crypto/bcrypt"
)

type RegisterUseCase struct {
	AuthRepository repositories.AuthRepository
}

func NewRegisterUseCase(repo repositories.AuthRepository) *RegisterUseCase {
	return &RegisterUseCase{AuthRepository: repo}
}

// Execute handles user registration. It returns the newly created user or an error.
// The responsibility of generating a JWT token is separated from this use case.
func (uc *RegisterUseCase) Execute(ctx context.Context, input dtos.RegisterInputDTO) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:    input.Email,
		Name:     input.Name,
		LastName: input.LastName,
		Username: input.UserName,
		Phone:    input.Phone,
	}

	// Register the user in the database
	newUser, err := uc.AuthRepository.Register(ctx, user)
	if err != nil {
		return nil, err
	}

	// Create the password entry for the new user
	if err := uc.AuthRepository.CreatePassword(ctx, newUser.ID, string(hashedPassword)); err != nil {
		return nil, err
	}

	// Update the last access time
	if err := uc.AuthRepository.SetLastAccess(ctx, newUser.ID); err != nil {
		// Note: Depending on business rules, this might not be a fatal error.
		// For now, we propagate it.
		return nil, err
	}

	return newUser, nil
}
