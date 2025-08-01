package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID int  `json:"user_id"`
	Role   Role `json:"role"`
	jwt.RegisteredClaims
}
