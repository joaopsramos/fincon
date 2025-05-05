package auth

import (
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/config"
	"github.com/lestrrat-go/jwx/jwa"
)

func GenerateJWTToken(userID uuid.UUID, expiresIn time.Duration) string {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{"sub": userID, "exp": time.Now().UTC().Add(expiresIn).Unix()},
	)

	tokenString, err := token.SignedString([]byte(config.Get().SecretKey))
	if err != nil {
		panic(err)
	}

	return tokenString
}

func NewTokenAuth() *jwtauth.JWTAuth {
	return jwtauth.New(jwa.HS256.String(), []byte(config.Get().SecretKey), nil)
}
