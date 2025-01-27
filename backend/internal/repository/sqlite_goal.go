package repository

import (
	"github.com/joaopsramos/fincon/internal/domain"
	"gorm.io/gorm"
)

type SQLiteGoalRepository struct {
	db *gorm.DB
}

func NewSQLiteGoal(db *gorm.DB) domain.GoalRepo {
	return SQLiteGoalRepository{db}
}

func (r SQLiteGoalRepository) All() []domain.Goal {
	var g []domain.Goal
	r.db.Find(&g)

	return g
}

func (r SQLiteGoalRepository) Get(id uint) (domain.Goal, error) {
	goal := domain.Goal{ID: id}
	result := r.db.Take(&goal)

	return goal, result.Error
}
