package user

import (
	"context"
	"fmt"
	auth "product-catalog/internal/auth"
	h "product-catalog/internal/dto"
	"product-catalog/internal/entity"
	custom "product-catalog/internal/errors"
	"time"
)

type Repository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id int) (*entity.User, error)
	UpdateByID(ctx context.Context, id int, user *entity.User) error
	DeleteByID(ctx context.Context, id int) error
	GetAll(ctx context.Context) ([]entity.User, error)
	ExistsByEmailOrUsername(ctx context.Context, email, username string) (bool, error)
}

type Service struct {
	repo   Repository
	hasher auth.Hasher
}

func NewUserService(repo Repository, hasher auth.Hasher) *Service {
	return &Service{repo: repo, hasher: hasher}
}

func (s *Service) CreateUser(ctx context.Context, input *h.CreateUserInput) error {
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

	newUser := &entity.User{
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

func (s *Service) UpdateUserByID(ctx context.Context, id int, userInfo *entity.User) error {
	err := s.repo.UpdateByID(ctx, id, userInfo)
	if err != nil {
		return fmt.Errorf("failed to update user by id: %w", err)
	}
	return nil
}

func (s *Service) DeleteUserByID(ctx context.Context, id int) error {
	err := s.repo.DeleteByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete user by id: %w", err)
	}
	return nil
}

func (s *Service) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	userFromDB, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return userFromDB, nil
}

func (s *Service) GetAllUsers(ctx context.Context) ([]entity.User, error) {
	users, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	return users, nil
}
