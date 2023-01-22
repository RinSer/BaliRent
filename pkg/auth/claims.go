package auth

import (
	"os"

	"github.com/golang-jwt/jwt/v4"
)

// jwtCustomClaims are custom claims extending default ones.
// See https://github.com/golang-jwt/jwt for more examples
type JwtCustomClaims struct {
	Id    int64  `json:"id"`
	Phone string `json:"phone"`
	Admin bool   `json:"admin"`
	jwt.RegisteredClaims
}

func GetJwtSecret() string {
	return os.Getenv("JWT_SECRET")
}
