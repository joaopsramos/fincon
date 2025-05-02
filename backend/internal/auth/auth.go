package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/config"
)

func GenerateJWTToken(userID uuid.UUID, expiresIn time.Duration) string {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{"sub": userID, "exp": time.Now().UTC().Add(expiresIn).Unix()},
	)

	tokenString, err := token.SignedString(config.SecretKey())
	if err != nil {
		panic(err)
	}

	return tokenString
}
