package repository

import (
	"github.com/joaopsramos/fincon/internal/domain"
	"gorm.io/gorm"
)

type PostgresGoalRepository struct {
	db *gorm.DB
}

func NewPostgresGoal(db *gorm.DB) domain.GoalRepo {
	return PostgresGoalRepository{db}
}

func (r PostgresGoalRepository) All() []domain.Goal {
	var g []domain.Goal
	r.db.Find(&g)

	return g
}

func (r PostgresGoalRepository) Get(id uint) (domain.Goal, error) {
	goal := domain.Goal{ID: id}
	result := r.db.Take(&goal)

	return goal, result.Error
}
