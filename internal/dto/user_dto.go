package dto

type CreateUserInput struct {
	Username string `validate:"required,min=3,max=12"`
	Password string `validate:"required,min=6,max=12"`
	Email    string `validate:"required,email"`
}
