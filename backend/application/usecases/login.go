package usecases

import (
	"backend/application/dtos"
	"backend/application/repositories"
	"backend/domain"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginUseCase struct {
	AuthRepository repositories.AuthRepository
}

func NewLoginUseCase(repo repositories.AuthRepository) *LoginUseCase {
	return &LoginUseCase{AuthRepository: repo}
}

func (uc *LoginUseCase) Execute(ctx *gin.Context, input dtos.LoginInputDTO) (*domain.User, error) {
	user, err := uc.AuthRepository.GetUserByEmail(ctx, &dtos.GetUserByEmailDTO{Email: input.Email})
	if err != nil {
		return nil, domain.ErrInternal
	}
	if user == nil {
		return nil, domain.ErrNotFound
	}

	hashedPassword, err := uc.AuthRepository.ValidateUserPasword(ctx, user.ID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(input.Password)); err != nil {
		return nil, domain.ErrInternal
	}

	return user, nil
}
