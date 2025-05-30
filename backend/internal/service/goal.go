package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/errs"
)

type GoalService struct {
	goalRepo domain.GoalRepo
}

type UpdateGoalDTO struct {
	ID         int
	Percentage int
}

func NewGoalService(goalRepo domain.GoalRepo) GoalService {
	return GoalService{goalRepo: goalRepo}
}

func (s *GoalService) All(ctx context.Context, userID uuid.UUID) []domain.Goal {
	return s.goalRepo.All(ctx, userID)
}

func (s *GoalService) Get(ctx context.Context, id uint, userID uuid.UUID) (*domain.Goal, error) {
	return s.goalRepo.Get(ctx, id, userID)
}

func (s *GoalService) Create(ctx context.Context, goals ...domain.Goal) error {
	return s.goalRepo.Create(ctx, goals...)
}

func (s *GoalService) UpdateAll(ctx context.Context, dtos []UpdateGoalDTO, userID uuid.UUID) ([]domain.Goal, error) {
	var zero []domain.Goal

	if len(dtos) < len(domain.DefaultGoalPercentages()) {
		return zero, errs.NewValidationError("one or more goals are missing")
	}

	percentageSum := 0
	for _, d := range dtos {
		if d.Percentage < 0 || d.Percentage > 100 {
			return zero, errs.NewValidationErrorF("invalid percentage for goal id %d, it must be between 1 and 100", d.ID)
		}

		percentageSum += d.Percentage
	}

	if percentageSum != 100 {
		return zero, errs.NewValidationError("the sum of all percentages must be equal to 100")
	}

	goals := s.All(ctx, userID)

	dtosByID := make(map[int]UpdateGoalDTO, len(dtos))
	for _, d := range dtos {
		dtosByID[d.ID] = d
	}

	for i, g := range goals {
		d, exists := dtosByID[int(g.ID)]
		if !exists {
			return zero, errs.NewValidationErrorF("missing goal with id %d", g.ID)
		}

		goals[i].Percentage = uint(d.Percentage)
	}

	err := s.goalRepo.UpdateAll(ctx, goals)

	return goals, err
}
