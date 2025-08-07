package http

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"product-catalog/internal/auth"
	"product-catalog/internal/domain"
	"product-catalog/internal/dto"
	custom "product-catalog/internal/errors"
	"strconv"
)

type UserService interface {
	CreateUser(ctx context.Context, user *dto.CreateUserInput) error
	GetUserByID(ctx context.Context, id int) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
	UpdateUserByID(ctx context.Context, requesterID, targetID int, input *dto.UpdateUserInput, role auth.Role) error
	DeleteUserByID(ctx context.Context, requesterID, targetID int, role auth.Role) error
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
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Get("/", h.GetAllUsers)
	r.Put("/{id}", h.UpdateUserByID)
	r.Delete("/{id}", h.DeleteUserByID)
	return r
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		h.logger.Warn("invalid method", zap.String("method", r.Method))
		return
	}

	var input dto.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("failed to decode request body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	defer r.Body.Close()

	err := h.svc.CreateUser(r.Context(), &input)
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

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.logger.Warn("invalid method", zap.String("method", r.Method))
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	role, ok := auth.RoleFromContext(r.Context())
	if !ok {
		h.logger.Warn("role not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}
	if role != auth.RoleAdmin {
		h.logger.Warn("unauthorized access")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}
	users, err := h.svc.GetAllUsers(r.Context())
	if err != nil {
		h.logger.Error("failed to get users", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) UpdateUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.logger.Warn("invalid method", zap.String("method", r.Method))
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role, ok := auth.RoleFromContext(r.Context())
	if !ok {
		h.logger.Warn("role not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	requesterID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		h.logger.Warn("user id not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	targetID, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Warn("invalid user id", zap.String("id", idStr), zap.Error(err))
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var input dto.UpdateUserInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("failed to decode input", zap.Error(err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if role != auth.RoleAdmin {
		input.Role = nil
	}

	err = h.svc.UpdateUserByID(r.Context(), requesterID, targetID, &input, role)
	if err != nil {
		h.logger.Error("failed to update user", zap.Error(err))
		if errors.Is(err, custom.ErrForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) DeleteUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.logger.Warn("invalid method", zap.String("method", r.Method))
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}

	role, ok := auth.RoleFromContext(r.Context())
	if !ok {
		h.logger.Warn("role not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	requesterID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		h.logger.Warn("user id not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}

	idStr := chi.URLParam(r, "id")
	targetID, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Warn("invalid user id", zap.String("id", idStr), zap.Error(err))
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	err = h.svc.DeleteUserByID(r.Context(), requesterID, targetID, role)
	if err != nil {
		h.logger.Error("failed to delete user", zap.Error(err))
		if errors.Is(err, custom.ErrForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		h.logger.Error("failed to delete user", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
