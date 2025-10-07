
package usecases

import (
	"backend/application/dtos"
	"backend/application/repositories"
	"backend/domain"
)

type RegisterUseCase struct {
	AuthRepository repositories.AuthRepository
}

func NewRegisterUseCase(repo repositories.AuthRepository) *RegisterUseCase {
	return &RegisterUseCase{AuthRepository: repo}
}

func (uc *RegisterUseCase) Execute(input dtos.RegisterInputDTO) (*domain.User, error) {
	user := &domain.User{
		Email:    input.Email,
		Name:     input.Name,
		LastName: input.LastName,
	}

	newUser, err := uc.AuthRepository.Register(user, input.Password)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return newUser, nil
}
