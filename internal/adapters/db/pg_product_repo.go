package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"product-catalog/internal/product"
)

type ProductRepo struct {
	db *pgxpool.Pool
}

func NewProductRepo(db *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) Create(ctx context.Context, product *product.Product) (int, error) {
	const query = `INSERT INTO products (title, price, description, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	var productID int
	err := r.db.QueryRow(ctx, query, product.Title, product.Price, product.Description, product.CreatedAt).Scan(&productID)
	if err != nil {
		return 0, fmt.Errorf("failed to create product: %w", err)
	}
	return productID, nil
}

func (r *ProductRepo) GetByID(ctx context.Context, id int) (*product.Product, error) {
	const query = `SELECT id, title, price, description, image_url, created_at FROM products WHERE id = $1`
	var productCard product.Product
	err := r.db.QueryRow(ctx, query, id).Scan(&productCard.ID, &productCard.Title, &productCard.Price, &productCard.Description, &productCard.Available, &productCard.ImageURL, &productCard.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get product by id: %w", err)
	}
	return &productCard, nil
}

func (r *ProductRepo) GetAll(ctx context.Context) ([]product.Product, error) {
	const query = `SELECT * FROM products`
	var products []product.Product
	row, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}
	defer row.Close()
	for row.Next() {
		var productCard product.Product
		err = row.Scan(&productCard.ID, &productCard.Title, &productCard.Price, &productCard.Description, &productCard.Available, &productCard.ImageURL, &productCard.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to get products: %w", err)
		}
		products = append(products, productCard)
	}
	return products, nil
}

func (r *ProductRepo) UpdateByID(ctx context.Context, id int, product *product.Product) error {
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
