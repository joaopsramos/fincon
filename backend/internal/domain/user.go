package domain

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
	GetByEmail(email string) (User, error)
}

var SecretKey = []byte(os.Getenv("SECRET_KEY"))

func (user *User) CreateToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Email,
		"exp": time.Now().UTC().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(SecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func EmailFromToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return SecretKey, nil
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
