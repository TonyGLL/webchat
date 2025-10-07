package dtos

type LoginInputDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterInputDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
	LastName string `json:"last_name" validate:"required"`
}

type GetUserByEmailDTO struct {
	Email string `json:"email"`
}
