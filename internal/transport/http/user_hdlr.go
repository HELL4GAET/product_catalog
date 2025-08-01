package http

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"product-catalog/internal/usecase/user"
)

type UserHandler struct {
	svc *user.Service
}

func NewUserHandler(svc *user.Service) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	return r
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
}
