package pg

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"product-catalog/internal/domain"
	custom "product-catalog/internal/errors"
)

type ProductRepo struct {
	db *pgxpool.Pool
}

func NewProductRepo(db *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) Create(ctx context.Context, product *domain.Product) (int, error) {
	const query = `INSERT INTO products (title, price, description, image_url, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var productID int
	err := r.db.QueryRow(ctx, query, product.Title, product.Price, product.Description, product.ImageURL, product.CreatedAt).Scan(&productID)
	if err != nil {
		return 0, fmt.Errorf("failed to create product: %w", err)
	}
	return productID, nil
}

func (r *ProductRepo) GetByID(ctx context.Context, id int) (*domain.Product, error) {
	const query = `SELECT id, title, price, description, available, image_url, created_at FROM products WHERE id = $1`
	var productCard domain.Product
	err := r.db.QueryRow(ctx, query, id).Scan(&productCard.ID, &productCard.Title, &productCard.Price, &productCard.Description, &productCard.Available, &productCard.ImageURL, &productCard.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, custom.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get product by id: %w", err)
	}
	return &productCard, nil
}

func (r *ProductRepo) GetAll(ctx context.Context) ([]domain.Product, error) {
	const query = `SELECT id, title, price, available, description, image_url, created_at FROM products ORDER BY id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get p: %w", err)
	}
	defer rows.Close()
	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		err = rows.Scan(&p.ID, &p.Title, &p.Price, &p.Available, &p.Description, &p.ImageURL, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *ProductRepo) UpdateByID(ctx context.Context, id int, product *domain.Product) error {
	const query = `UPDATE products SET title = $1, price = $2, description = $3, available = $4 WHERE id = $5`
	_, err := r.db.Exec(ctx, query, product.Title, product.Price, product.Description, product.Available, id)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	return nil
}

func (r *ProductRepo) DeleteByID(ctx context.Context, id int) error {
	const query = `DELETE FROM products WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	return nil
}
