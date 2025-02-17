package repository

import (
	"context"

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

func (r PostgresGoalRepository) All(ctx context.Context, userID uuid.UUID) []domain.Goal {
	var g []domain.Goal
	r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&g)

	return g
}

func (r PostgresGoalRepository) Get(ctx context.Context, id uint, userID uuid.UUID) (*domain.Goal, error) {
	goal := domain.Goal{ID: id}

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Take(&goal).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &domain.Goal{}, errs.NewNotFound("goal")
		}

		return &domain.Goal{}, err
	}

	return &goal, nil
}

func (r PostgresGoalRepository) Create(ctx context.Context, goals ...domain.Goal) error {
	if err := r.db.WithContext(ctx).Create(goals).Error; err != nil {
		return err
	}

	return nil
}

func (r PostgresGoalRepository) UpdateAll(ctx context.Context, goals []domain.Goal) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, g := range goals {
			if err := r.db.Save(&g).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return err
}
