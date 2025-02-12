package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/domain"
	errs "github.com/joaopsramos/fincon/internal/error"
	"gorm.io/gorm"
)

type PostgresSalaryRepository struct {
	db *gorm.DB
}

func NewPostgresSalary(db *gorm.DB) domain.SalaryRepo {
	return PostgresSalaryRepository{db}
}

func (r PostgresSalaryRepository) Get(userID uuid.UUID) (*domain.Salary, error) {
	var s domain.Salary

	err := r.db.Where("user_id = ?", userID).Take(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &domain.Salary{}, errs.NewNotFound("salary")
	} else if err != nil {
		return &domain.Salary{}, err
	}

	return &s, nil
}

func (r PostgresSalaryRepository) Create(s *domain.Salary) error {
	if err := r.db.Create(s).Error; err != nil {
		return err
	}

	return nil
}

func (r PostgresSalaryRepository) Update(s *domain.Salary) error {
	if err := r.db.Model(s).Updates(*s).Error; err != nil {
		return err
	}

	return nil
}
