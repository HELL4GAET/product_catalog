package auth

import "fmt"

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

type bcryptHasher struct{}

func NewHasher() Hasher {
	return &bcryptHasher{}
}

func (h *bcryptHasher) Hash(password string) (string, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return "", fmt.Errorf("hasher: hash password: %w", err)
	}
	return hash, nil
}

func (h *bcryptHasher) Compare(hashedPassword, password string) error {
	if err := ComparePassword(hashedPassword, password); err != nil {
		return fmt.Errorf("hasher: compare password: %w", err)
	}
	return nil
}
