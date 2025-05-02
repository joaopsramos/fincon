package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string    `gorm:"type:citext"`
	HashPassword string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserDTO struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func (u *User) ToDTO() UserDTO {
	return UserDTO{
		ID:    u.ID,
		Email: u.Email,
	}
}

type UserRepo interface {
	Create(ctx context.Context, user *User, salary *Salary) error
	Get(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}
