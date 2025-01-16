package repository

import (
	"github.com/joaopsramos/fincon/internal/domain"
	"gorm.io/gorm"
)

type SQLiteExpenseRepository struct {
	db *gorm.DB
}

func NewSQLiteExpense(db *gorm.DB) domain.ExpenseRepository {
	return SQLiteExpenseRepository{db}
}

func (r SQLiteExpenseRepository) GetByGoalID(goalID uint, conditions ...any) []domain.Expense {
	var e []domain.Expense
	r.db.
		Where("goal_id = ?", goalID).
		Where(conditions).
		Order("date DESC, created_at DESC").
		Find(&e)

	return e
}
