package dtos

// RefreshTokenInputDTO represents the data required to refresh an authentication token.
type RefreshTokenInputDTO struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}