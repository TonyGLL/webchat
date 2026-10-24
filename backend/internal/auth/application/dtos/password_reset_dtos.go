package dtos

type ForgotPasswordInputDTO struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

type ResetPasswordInputDTO struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}
