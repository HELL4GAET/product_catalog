package user

import (
	"context"
	"fmt"
	auth "product-catalog/internal/auth"
	"product-catalog/internal/domain"
	"product-catalog/internal/dto"
	custom "product-catalog/internal/errors"
	"time"
)

type Repository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int) (*domain.User, error)
	UpdateByID(ctx context.Context, id int, username, email string, role auth.Role, passwordHash string) error
	DeleteByID(ctx context.Context, id int) error
	GetAll(ctx context.Context) ([]domain.User, error)
	ExistsByEmailOrUsername(ctx context.Context, email, username string) (bool, error)
	GetUserCredsAndRoleByEmail(ctx context.Context, email string) (string, int, auth.Role, error)
}

type JwtService interface {
	GenerateToken(userID int, role auth.Role) (string, error)
	ParseToken(tokenStr string) (*auth.JWTClaims, error)
}

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

type Service struct {
	repo   Repository
	hasher Hasher
	jwtSvc JwtService
}

func NewUserService(repo Repository, hasher Hasher, jwtSvc JwtService) *Service {
	return &Service{repo: repo, hasher: hasher, jwtSvc: jwtSvc}
}

func (s *Service) CreateUser(ctx context.Context, input *dto.CreateUserInput) error {
	exists, err := s.repo.ExistsByEmailOrUsername(ctx, input.Email, input.Username)
	if err != nil {
		return fmt.Errorf("failed to check user exists: %w", err)
	}
	if exists {
		return custom.ErrConflict
	}

	hash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &domain.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: hash,
		CreatedAt:    time.Now(),
		Role:         auth.RoleUser,
	}

	err = s.repo.Create(ctx, newUser)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (s *Service) UpdateUserByID(ctx context.Context, requesterID, targetID int, input *dto.UpdateUserInput, role auth.Role) error {
	if role != auth.RoleAdmin && requesterID != targetID {
		return custom.ErrForbidden
	}

	userFromDB, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	newRole := userFromDB.Role
	if role == auth.RoleAdmin && input.Role != nil {
		newRole = *input.Role
	}

	username := userFromDB.Username
	if input.Username != nil {
		username = *input.Username
	}

	email := userFromDB.Email
	if input.Email != nil {
		email = *input.Email
	}

	passwordHash := userFromDB.PasswordHash
	if input.Password != nil {
		newPasswordHash, err := s.hasher.Hash(*input.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		passwordHash = newPasswordHash
	}
	err = s.repo.UpdateByID(ctx, targetID, username, email, newRole, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (s *Service) DeleteUserByID(ctx context.Context, requesterID, targetID int, role auth.Role) error {
	if role != auth.RoleAdmin && requesterID != targetID {
		return custom.ErrForbidden
	}
	err := s.repo.DeleteByID(ctx, targetID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (s *Service) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
	userFromDB, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return userFromDB, nil
}

func (s *Service) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	users, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	return users, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	pwHash, id, role, err := s.repo.GetUserCredsAndRoleByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("failed to check user exists: %w", err)
	}

	err = s.hasher.Compare(pwHash, password)
	if err != nil {
		return "", custom.ErrUnauthorized
	}

	token, err := s.jwtSvc.GenerateToken(id, role)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return token, nil
}
