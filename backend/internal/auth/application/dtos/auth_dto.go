package dtos

type LoginInputDTO struct {
	User     string `json:"user" validate:"required,max=255"`
	Password string `json:"password" validate:"required,max=72"`
	// IP is the client address, filled by the controller for per-account throttling.
	IP string `json:"-"`
}

type RegisterInputDTO struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Name     string `json:"name" validate:"required,max=100"`
	LastName string `json:"last_name" validate:"required,max=100"`
	Username string `json:"username" validate:"required,min=3,max=30"` // charset: domain.ValidateUsername
	Phone    string `json:"phone" validate:"required,max=30"`
}

type SendVerifyEmailInputDTO struct {
	Email string `form:"email" validate:"required,email"`
	ID    int    `uri:"user_id" validate:"required,gt=0"`
}

type VerifyEmailInputDTO struct {
	Token string `uri:"token" validate:"required"`
}
