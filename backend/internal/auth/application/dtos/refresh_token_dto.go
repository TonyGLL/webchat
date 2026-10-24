package dtos

// RefreshTokenInputDTO is the optional body of a refresh request; browsers send the
// token in the httpOnly cookie instead.
type RefreshTokenInputDTO struct {
	RefreshToken string `json:"refreshToken"`
}
