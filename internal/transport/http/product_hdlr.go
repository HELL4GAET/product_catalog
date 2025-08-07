package http

import (
	"context"
	"github.com/go-chi/chi/v5"
	"net/http"
	"product-catalog/internal/domain"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product *domain.Product) (int, error)
	GetProductByID(ctx context.Context, id int) (*domain.Product, error)
	GetAllProducts(ctx context.Context) ([]domain.Product, error)
	UpdateProductByID(ctx context.Context, id int, product *domain.Product) error
	DeleteProductByID(ctx context.Context, id int) error
}

type ProductHandler struct {
	svc ProductService
}

func NewProductHandler(svc ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func (h *ProductHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.CreateProduct)
	return r
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
