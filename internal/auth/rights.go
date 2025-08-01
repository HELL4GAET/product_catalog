package auth

import (
	"product-catalog/internal/errors"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (r Role) CanDoAdminAction() error {
	if r != RoleAdmin {
		return errors.ErrForbidden
	}
	return nil
}
