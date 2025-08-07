package dto

import "product-catalog/internal/auth"

type CreateUserInput struct {
	Username string `json:"username" validate:"jsonrequired,min=3,max=12"`
	Password string `json:"password" validate:"required,min=6,max=12"`
	Email    string `json:"email" validate:"required,email"`
}

type LoginInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=6,max=12"`
}

type UpdateUserInput struct {
	Username *string    `json:"username" validate:"jsonrequired,min=3,max=12"`
	Email    *string    `json:"email" validate:"required,email"`
	Password *string    `json:"password" validate:"required,min=6,max=12"`
	Role     *auth.Role `json:"role,omitempty"`
}
