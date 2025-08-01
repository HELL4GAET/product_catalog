package http

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"product-catalog/internal/usecase/product"
)

type ProductHandler struct {
	svc *product.Service
}

func NewProductHandler(svc *product.Service) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func (h *ProductHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	return r
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {

}
