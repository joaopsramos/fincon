package repository

import (
	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/domain"
	"gorm.io/gorm"
)

type PostgresSalaryRepository struct {
	db *gorm.DB
}

func NewPostgresSalary(db *gorm.DB) domain.SalaryRepo {
	return PostgresSalaryRepository{db}
}

func (r PostgresSalaryRepository) Get(userID uuid.UUID) domain.Salary {
	var s domain.Salary
	r.db.Where("user_id = ?", userID).Take(&s)
	return s
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
