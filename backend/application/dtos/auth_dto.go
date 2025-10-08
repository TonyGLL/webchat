package dtos

import "backend/domain"

type LoginInputDTO struct {
	User     string `json:"user" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RegisterInputDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
	LastName string `json:"last_name" validate:"required"`
	UserName string `json:"username" validate:"required,min=3,max=30"`
	Phone    string `json:"phone" validate:"required"`
}

type GetUserByEmailOrUsernameDTO struct {
	User string `json:"user"`
}

type LoginResponseDTO struct {
	User  domain.User `json:"user"`
	Token string      `json:"token"`
}

type RegisterResponseDTO struct {
	ID    int    `json:"id"`
	Token string `json:"token"`
}
