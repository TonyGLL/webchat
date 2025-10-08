package dtos

type LoginInputDTO struct {
	User     string `json:"user" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RegisterInputDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
	LastName string `json:"last_name" validate:"required"`
	Username string `json:"username" validate:"required,min=3,max=30"`
	Phone    string `json:"phone" validate:"required"`
}
