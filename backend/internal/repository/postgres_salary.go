package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/errs"
	"gorm.io/gorm"
)

type PostgresSalaryRepository struct {
	db *gorm.DB
}

func NewPostgresSalary(db *gorm.DB) domain.SalaryRepo {
	return PostgresSalaryRepository{db}
}

func (r PostgresSalaryRepository) Get(ctx context.Context, userID uuid.UUID) (*domain.Salary, error) {
	var s domain.Salary

	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Take(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &domain.Salary{}, errs.NewNotFound("salary")
	} else if err != nil {
		return &domain.Salary{}, err
	}

	return &s, nil
}

func (r PostgresSalaryRepository) Create(ctx context.Context, s *domain.Salary) error {
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		return err
	}

	return nil
}

func (r PostgresSalaryRepository) Update(ctx context.Context, s *domain.Salary) error {
	if err := r.db.WithContext(ctx).Model(s).Updates(*s).Error; err != nil {
		return err
	}

	return nil
}
