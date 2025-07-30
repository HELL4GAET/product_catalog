package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func (s *Storage) New(ctx context.Context, connectionString string) (*Storage, error) {
	dbPool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return &Storage{db: dbPool}, nil
}

func (s *Storage) DB() *pgxpool.Pool { return s.db }

func (s *Storage) Close() { s.db.Close() }
