package pg

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"product-catalog/internal/auth"
	"product-catalog/internal/domain"
	custom "product-catalog/internal/errors"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	const query = `INSERT INTO users (username, email, password_hash, role, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, user.Username, user.Email, user.PasswordHash, user.Role, user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return custom.ErrConflict
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int) (*domain.User, error) {
	const query = `SELECT id, username, email, role, created_at FROM users WHERE id = $1`
	var userFromDB domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(&userFromDB.ID, &userFromDB.Username, &userFromDB.Email, &userFromDB.Role, &userFromDB.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, custom.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &userFromDB, nil
}

func (r *UserRepo) GetAll(ctx context.Context) ([]domain.User, error) {
	const query = `SELECT * FROM users`
	var users []domain.User
	row, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer row.Close()
	for row.Next() {
		var userFromDB domain.User
		err = row.Scan(&userFromDB.ID, &userFromDB.Username, &userFromDB.Email, &userFromDB.Role, &userFromDB.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to get users: %w", err)
		}
		users = append(users, userFromDB)
	}
	return users, nil
}

func (r *UserRepo) UpdateByID(ctx context.Context, id int, username, email string, role auth.Role) error {
	const query = `UPDATE users SET username = $1, email = $2, role = $3 WHERE id = $4`
	_, err := r.db.Exec(ctx, query, username, email, role, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return custom.ErrConflict
		}
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *UserRepo) DeleteByID(ctx context.Context, id int) error {
	const query = `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return custom.ErrNotFound
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (r *UserRepo) ExistsByEmailOrUsername(ctx context.Context, email, username string) (bool, error) {
	const query = `
		SELECT 1 FROM users 
		WHERE email = $1 OR username = $2 
		LIMIT 1
	`
	var exists int
	err := r.db.QueryRow(ctx, query, email, username).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}
	return true, nil
}

func (r *UserRepo) GetUserCredsAndRoleByEmail(ctx context.Context, email string) (string, int, auth.Role, error) {
	const query = `SELECT password_hash, id, role FROM users WHERE email = $1`

	var pwHash string
	var id int
	var role auth.Role
	err := r.db.QueryRow(ctx, query, email).Scan(&pwHash, &id, &role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", 0, "", custom.ErrNotFound
		}
		return "", 0, "", fmt.Errorf("failed to check if user exists: %w", err)
	}
	return pwHash, id, role, nil
}
