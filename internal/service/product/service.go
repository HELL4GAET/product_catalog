package product

import (
	"context"
	"fmt"
	"product-catalog/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, product *domain.Product) (int, error)
	GetByID(ctx context.Context, id int) (*domain.Product, error)
	GetAll(ctx context.Context) ([]domain.Product, error)
	UpdateByID(ctx context.Context, id int, product *domain.Product) error
	DeleteByID(ctx context.Context, id int) error
}

type Service struct {
	repo Repository
}

func NewProductService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) CreateProduct(ctx context.Context, product *domain.Product) (int, error) {
	id, err := s.repo.Create(ctx, product)
	if err != nil {
		return 0, fmt.Errorf("failed to create product: %w", err)
	}
	return id, nil
}

func (s *Service) GetProductByID(ctx context.Context, id int) (*domain.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product by id: %w", err)
	}
	return product, nil
}

func (s *Service) GetAllProducts(ctx context.Context) ([]domain.Product, error) {
	product, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all products: %w", err)
	}
	return product, nil
}

func (s *Service) UpdateProductByID(ctx context.Context, id int, product *domain.Product) error {
	err := s.repo.UpdateByID(ctx, id, product)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	return nil
}

func (s *Service) DeleteProductByID(ctx context.Context, id int) error {
	return s.repo.DeleteByID(ctx, id)
}
