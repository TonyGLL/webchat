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
}

func NewRegisterUseCase(
	authRepo domain.AuthRepository,
	ps domain.PasswordService,
	jwtS shared_app.JwtService) *RegisterUseCase {
	return &RegisterUseCase{
		authRepo:        authRepo,
		passwordService: ps,
		jwtService:      jwtS,
	}
}

// Execute handles user registration; the user and password are created atomically by the repository.
// It returns the newly created user or an error.
func (uc *RegisterUseCase) Execute(ctx context.Context, input dtos.RegisterInputDTO) (*dtos.UserResponseDTO, error) {
	if err := domain.ValidatePassword(input.Password, input.Username, input.Email); err != nil {
		return nil, err
	}

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

	createdUser, err := uc.authRepo.RegisterWithPassword(ctx, user, hashedPassword)
	if err != nil {
		return nil, err
	}

	response := dtos.UserResponseDTO{
		ID:    createdUser.ID,
		Email: createdUser.Email,
	}

	return &response, nil
}
