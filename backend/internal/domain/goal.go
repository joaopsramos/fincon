package domain

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

	Expenses []Expense `json:"-"`
}

type GoalRepo interface {
	All() []Goal
	Get(id uint) (Goal, error)
}
