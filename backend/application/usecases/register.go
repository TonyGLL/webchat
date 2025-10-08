package usecases

import (
	"backend/application/dtos"
	"backend/application/repositories"
	"backend/domain"
	"backend/infrastructure/jwt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUseCase struct {
	AuthRepository repositories.AuthRepository
}

func NewRegisterUseCase(repo repositories.AuthRepository) *RegisterUseCase {
	return &RegisterUseCase{AuthRepository: repo}
}

func (uc *RegisterUseCase) Execute(ctx *gin.Context, input dtos.RegisterInputDTO) (*dtos.RegisterResponseDTO, error) {
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

	newUser, err := uc.AuthRepository.Register(ctx, user)
	if err != nil {
		return nil, err
	}

	if err := uc.AuthRepository.CreatePassword(ctx, newUser.ID, string(hashedPassword)); err != nil {
		return nil, err
	}

	if err := uc.AuthRepository.SetLastAccess(ctx, newUser.ID); err != nil {
		return nil, domain.ErrInternal
	}

	jwtImpl := jwt.NewJWTService()
	token, err := jwtImpl.GenerateToken(newUser.ID, 24*time.Hour)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return nil, domain.ErrInternal
	}

	response := dtos.RegisterResponseDTO{
		ID:    newUser.ID,
		Token: token,
	}

	return &response, nil
}
