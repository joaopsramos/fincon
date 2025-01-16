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
	ID         uint `gorm:"primaryKey"`
	Name       GoalName
	Percentage uint

	Expenses []Expense
}
