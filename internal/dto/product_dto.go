package dto

type CreateProductInput struct {
	Title       string `validate:"required,min=3,max=12"`
	Price       int    `validate:"required,min=1,max=100000"`
	Description string `validate:"required,min=3,max=120"`
	Available   bool   `validate:"required"`
	ImageURL    string `validate:"required,url"`
}
