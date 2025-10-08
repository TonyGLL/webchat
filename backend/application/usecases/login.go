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

type LoginUseCase struct {
	AuthRepository repositories.AuthRepository
}

func NewLoginUseCase(repo repositories.AuthRepository) *LoginUseCase {
	return &LoginUseCase{AuthRepository: repo}
}

func (uc *LoginUseCase) Execute(ctx *gin.Context, input dtos.LoginInputDTO) (*dtos.LoginResponseDTO, error) {
	user, err := uc.AuthRepository.GetUserByEmailOrUsername(ctx, &dtos.GetUserByEmailOrUsernameDTO{User: input.User})
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

	if err := uc.AuthRepository.SetLastAccess(ctx, user.ID); err != nil {
		return nil, domain.ErrInternal
	}

	jwtImpl := jwt.NewJWTService()
	token, err := jwtImpl.GenerateToken(user.ID, 24*time.Hour)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return nil, domain.ErrInternal
	}

	response := dtos.LoginResponseDTO{
		User:  *user,
		Token: token,
	}

	return &response, nil
}
