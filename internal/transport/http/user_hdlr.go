package http

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"product-catalog/internal/dto"
	"product-catalog/internal/entity"
	custom "product-catalog/internal/errors"
)

var ctx = context.Background()

type UserService interface {
	CreateUser(ctx context.Context, user *dto.CreateUserInput) error
	GetUserByID(ctx context.Context, id int) (*entity.User, error)
	GetAllUsers(ctx context.Context) ([]entity.User, error)
	UpdateUserByID(ctx context.Context, id int, user *entity.User) error
	DeleteUserByID(ctx context.Context, id int) error
	Login(ctx context.Context, email, password string) (string, error)
}

type UserHandler struct {
	svc    UserService
	logger *zap.Logger
}

func NewUserHandler(svc UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{svc: svc, logger: logger}
}

func (h *UserHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Register)
	return r
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		h.logger.Warn("invalid method", zap.String("method", r.Method))
		return
	}

	input := dto.CreateUserInput{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
		Email:    r.FormValue("email"),
	}

	err := h.svc.CreateUser(ctx, &input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error("failed to create user", zap.Error(err), zap.String("email", input.Email), zap.String("username", input.Username))
		return
	}

	h.logger.Info("user created successfully", zap.String("email", input.Email), zap.String("username", input.Username))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(`{"message": "user created"}`))
	if err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.Warn("invalid method", zap.String("method", r.Method))
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	input := dto.LoginInput{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	h.logger.Info("login attempt", zap.String("email", input.Email))

	token, err := h.svc.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		h.logger.Warn("unauthorized login", zap.String("email", input.Email), zap.Error(err))
		http.Error(w, custom.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	h.logger.Info("login successful", zap.String("email", input.Email))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"token": token})
}
