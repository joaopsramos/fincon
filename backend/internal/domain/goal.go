package domain

import "github.com/google/uuid"

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
	ID         uint     `json:"id" gorm:"primaryKey"`
	Name       GoalName `json:"name"`
	Percentage uint     `json:"percentage"`

	UserID uuid.UUID `json:"-" gorm:"type:uuid"`

	Expenses []Expense `json:"-"`
}

type GoalRepo interface {
	All(userID uuid.UUID) []Goal
	Get(id uint, userID uuid.UUID) (Goal, error)
}
