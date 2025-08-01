package entity

import (
	"product-catalog/internal/auth"
	"time"
)

type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	Role         auth.Role
	CreatedAt    time.Time
}
