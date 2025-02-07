package repository

import (
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/domain"
	"gorm.io/gorm"
)

type PostgresGoalRepository struct {
	db *gorm.DB
}

func NewPostgresGoal(db *gorm.DB) domain.GoalRepo {
	return PostgresGoalRepository{db}
}

func (r PostgresGoalRepository) All(userID uuid.UUID) []domain.Goal {
	var g []domain.Goal
	r.db.Where("user_id = ?", userID).Find(&g)

	return g
}

func (r PostgresGoalRepository) Get(id uint, userID uuid.UUID) (domain.Goal, error) {
	goal := domain.Goal{ID: id}
	result := r.db.Where("user_id = ?", userID).Take(&goal)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return domain.Goal{}, errors.New("goal not found")
		}

		slog.Error(result.Error.Error())
		return domain.Goal{}, errors.New("goal could not be retrieved")
	}

	return goal, nil
}
