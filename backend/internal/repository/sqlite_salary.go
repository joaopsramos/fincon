package repository

import (
	"github.com/joaopsramos/fincon/internal/domain"
	"gorm.io/gorm"
)

type PostgresSalaryRepository struct {
	db *gorm.DB
}

func NewPostgresSalary(db *gorm.DB) domain.SalaryRepo {
	return PostgresSalaryRepository{db}
}

func (r PostgresSalaryRepository) Get() domain.Salary {
	var s domain.Salary
	r.db.First(&s)
	return s
}
