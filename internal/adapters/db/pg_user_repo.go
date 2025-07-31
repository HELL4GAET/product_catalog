package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"product-catalog/internal/user"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *user.User) error {
	const query = `INSERT INTO users (username, email, password_hash, role, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, user.Username, user.Email, user.PasswordHash, user.Role, user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int) (*user.User, error) {
	const query = `SELECT id, username, email, role, created_at FROM users WHERE id = $1`
	var userFromDB user.User
	err := r.db.QueryRow(ctx, query, id).Scan(&userFromDB.ID, &userFromDB.Username, &userFromDB.Email, &userFromDB.Role, &userFromDB.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &userFromDB, nil
}

func (r *UserRepo) GetAll(ctx context.Context) ([]user.User, error) {
	const query = `SELECT * FROM users`
	var users []user.User
	row, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer row.Close()
	for row.Next() {
		var userFromDB user.User
		err = row.Scan(&userFromDB.ID, &userFromDB.Username, &userFromDB.Email, &userFromDB.Role, &userFromDB.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to get users: %w", err)
		}
		users = append(users, userFromDB)
	}
	return users, nil
}

func (r *UserRepo) UpdateByID(ctx context.Context, id int, user *user.User) error {
	const query = `UPDATE users SET username = $1, email = $2, role = $3 WHERE id = $4`
	_, err := r.db.Exec(ctx, query, user.Username, user.Email, user.Role, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
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
			return ErrNotFound
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
