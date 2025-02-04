package repository

import (
	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/domain"
	"gorm.io/gorm"
)

type PostgresUserRepository struct {
	db *gorm.DB
}

func NewPostgresUser(db *gorm.DB) domain.UserRepo {
	return PostgresUserRepository{db}
}

func (r PostgresUserRepository) Create(user *domain.User) error {
	result := r.db.Create(user)
	return result.Error
}

func (r PostgresUserRepository) Get(id uuid.UUID) (domain.User, error) {
	var user domain.User
	result := r.db.Take(&user, id)
	return user, result.Error
}

func (r PostgresUserRepository) GetByEmail(email string) (domain.User, error) {
	var user domain.User
	result := r.db.Where("email = ?", email).Take(&user)
	return user, result.Error
}
