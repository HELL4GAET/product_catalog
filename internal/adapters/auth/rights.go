package auth

import (
	"product-catalog/internal/adapters/db"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (r Role) CanDoAdminAction() error {
	if r != RoleAdmin {
		return db.ErrForbidden
	}
	return nil
}
