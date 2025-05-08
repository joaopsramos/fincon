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

type UserToken struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	Token     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	ExpiresAt time.Time
	Used      bool

	CreatedAt time.Time
	UpdatedAt time.Time

	User User `gorm:"foreignKey:UserID"`
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
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	CreateToken(ctx context.Context, token *UserToken) error
	GetUserTokenByToken(ctx context.Context, token string) (*UserToken, error)
	MarkTokenAsUsed(ctx context.Context, tokenID uint) error
}
