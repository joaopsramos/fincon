package domain

import (
	"context"

	"github.com/google/uuid"
)

type GoalName string

const (
	FixedCosts           GoalName = "Fixed costs"
	Comfort              GoalName = "Comfort"
	Goals                GoalName = "Goals"
	Pleasures            GoalName = "Pleasures"
	FinancialInvestments GoalName = "Financial investments"
	Knowledge            GoalName = "Knowledge"
)

type Goal struct {
	ID         uint
	Name       GoalName
	Percentage uint
	UserID     uuid.UUID

	Expenses []Expense
}

type GoalDTO struct {
	ID         uint     `json:"id"`
	Name       GoalName `json:"name"`
	Percentage uint     `json:"percentage"`
}

func (g *Goal) ToDTO() GoalDTO {
	return GoalDTO{
		ID:         g.ID,
		Name:       g.Name,
		Percentage: g.Percentage,
	}
}

func DefaultGoalPercentages() map[GoalName]uint {
	return map[GoalName]uint{
		FixedCosts:           40,
		Comfort:              20,
		Goals:                5,
		Pleasures:            5,
		FinancialInvestments: 25,
		Knowledge:            5,
	}
}

type GoalRepo interface {
	All(ctx context.Context, userID uuid.UUID) []Goal
	Get(ctx context.Context, id uint, userID uuid.UUID) (*Goal, error)
	Create(ctx context.Context, goals ...Goal) error
	UpdateAll(ctx context.Context, goals []Goal) error
}
