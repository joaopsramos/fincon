package domain

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/config"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string    `json:"email" gorm:"type:citext"`
	HashPassword string    `json:"-"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type UserRepo interface {
	Create(user *User) error
	Get(id uuid.UUID) (User, error)
	GetByEmail(email string) (User, error)
}

func CreateToken(userID uuid.UUID, expiresIn time.Duration) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().UTC().Add(expiresIn).Unix(),
	})

	tokenString, err := token.SignedString(config.SecretKey())
	if err != nil {
		panic(err)
	}

	return tokenString
}

func UserIDFromToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return config.SecretKey(), nil
	}, jwt.WithValidMethods([]string{"alg"}))
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	sub, err := token.Claims.GetSubject()
	if err != nil {
		panic(err)
	}

	return sub, nil
}
