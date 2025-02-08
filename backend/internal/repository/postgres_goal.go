package repository

import (
	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/domain"
	errs "github.com/joaopsramos/fincon/internal/error"
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

	if err := r.db.Where("user_id = ?", userID).Take(&goal).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.Goal{}, errs.NewNotFound("goal")
		}

		return domain.Goal{}, err
	}

	return goal, nil
}

func (r PostgresGoalRepository) Create(goals ...domain.Goal) error {
	if err := r.db.Create(goals).Error; err != nil {
		return err
	}

	return nil
}
