package repository

import (
	"github.com/joaopsramos/fincon/internal/domain"
	"gorm.io/gorm"
)

type SQLiteSalaryRepository struct {
	db *gorm.DB
}

func NewSQLiteSalary(db *gorm.DB) domain.SalaryRepo {
	return SQLiteSalaryRepository{db}
}

func (r SQLiteSalaryRepository) Get() domain.Salary {
	var s domain.Salary
	r.db.First(&s)
	return s
}
