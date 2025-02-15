package service_test

import (
	"testing"
	"time"

	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/repository"
	"github.com/joaopsramos/fincon/internal/service"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func NewTestExpenseService(t *testing.T, tx *gorm.DB) service.ExpenseService {
	t.Helper()

	salaryRepo := repository.NewPostgresSalary(tx)
	goalRepo := repository.NewPostgresGoal(tx)
	expenseRepo := repository.NewPostgresExpense(tx)

	return service.NewExpenseService(expenseRepo, goalRepo, salaryRepo)
}

func TestPostgresExpense_GetSummary(t *testing.T) {
	t.Parallel()
	tx := testhelper.NewTestPostgresTx(t)
	f := testhelper.NewFactory(tx)
	user := f.InsertUser()

	salary := domain.Salary{Amount: 10_000 * 100, UserID: user.ID}
	f.InsertSalary(&salary)

	goalsByName := make(map[domain.GoalName]domain.Goal)

	goalsToInsert := []struct {
		name       domain.GoalName
		percentage uint
	}{
		{domain.Comfort, 20},              // limit 2000
		{domain.FixedCosts, 40},           // limit 4000
		{domain.Goals, 5},                 // limit 500
		{domain.Pleasures, 5},             // limit 500
		{domain.FinancialInvestments, 25}, // limit 2500
		{domain.Knowledge, 5},             // limit 500
	}
	for _, g := range goalsToInsert {
		goal := domain.Goal{Name: g.name, Percentage: g.percentage, UserID: user.ID}
		f.InsertGoal(&goal)
		goalsByName[goal.Name] = goal
	}

	now := testhelper.MiddleOfMonth()
	oneMonthAgo := now.AddDate(0, -1, 0)
	nextMonth := now.AddDate(0, 1, 0)

	expenses := []struct {
		value    float64
		date     time.Time
		goalName domain.GoalName
	}{
		{125.74, oneMonthAgo, domain.Comfort},
		{500, oneMonthAgo, domain.Pleasures},

		{50.5, now, domain.Comfort},
		{125.49, now, domain.Comfort},
		{400, now, domain.FixedCosts},
		{100, now, domain.FixedCosts},
		{190.89, now, domain.Pleasures},
		{340.15, now, domain.Pleasures},
		{900.99, now, domain.Knowledge},

		{10, nextMonth, domain.Pleasures},
		{700.25, nextMonth, domain.FinancialInvestments},
	}

	for _, e := range expenses {
		f.InsertExpense(&domain.Expense{
			Value:     int64(e.value * 100),
			Date:      e.date,
			GoalID:    goalsByName[e.goalName].ID,
			CreatedAt: now,
		})
	}

	type expectedEntries struct {
		goalName  domain.GoalName
		spent     float64
		mustSpend float64
		used      float64
		total     float64
	}

	assertSummaryGoals := func(a *assert.Assertions, tests []expectedEntries, entries []service.SummaryGoal) {
		entriesByName := make(map[domain.GoalName]service.SummaryGoal)
		for _, e := range entries {
			entriesByName[domain.GoalName(e.Name)] = e
		}

		for _, tt := range tests {
			entry := entriesByName[tt.goalName]
			a.Equal(string(tt.goalName), entry.Name)
			a.Equal(tt.spent, entry.Spent)
			a.Equal(tt.mustSpend, entry.MustSpend)
			a.Equal(tt.used, float64(int(entry.Used*100))/100)
			a.Equal(tt.total, float64(int(entry.Total*100))/100)
		}
	}

	twoMonthsAgo := now.AddDate(0, -2, 0)

	tests := []struct {
		name      string
		date      time.Time
		entries   []expectedEntries
		spent     float64
		mustSpend float64
		used      float64
	}{
		{
			name: "should return zero values for two months ago",
			date: twoMonthsAgo,
			entries: []expectedEntries{
				{domain.Comfort, 0, 2000, 0, 0},
				{domain.FixedCosts, 0, 4000, 0, 0},
				{domain.Pleasures, 0, 500, 0, 0},
				{domain.Knowledge, 0, 500, 0, 0},
				{domain.Goals, 0, 500, 0, 0},
				{domain.FinancialInvestments, 0, 2500, 0, 0},
			},
			spent:     0.0,
			mustSpend: 10000.0,
			used:      0.0,
		},
		{
			name: "should show previous month spending and limits",
			date: oneMonthAgo,
			entries: []expectedEntries{
				{domain.Comfort, 125.74, 2000, 6.28, 1.25},
				{domain.Comfort, 125.74, 2000, 6.28, 1.25},
				{domain.FixedCosts, 0, 4000, 0.0, 0.0},
				{domain.Pleasures, 500, 500, 100.0, 5.0},
				{domain.Knowledge, 0, 500, 0.0, 0.0},
				{domain.Goals, 0, 500, 0, 0},
				{domain.FinancialInvestments, 0, 2500, 0.0, 0.0},
			},
			spent:     625.74,
			mustSpend: 9374.26,
			used:      6.2574,
		},
		{
			name: "should calculate current month totals with proper limits",
			date: now,
			entries: []expectedEntries{
				{domain.Comfort, 50.5 + 125.49, 2000, 8.79, 1.75},
				{domain.FixedCosts, 400 + 100, 4000, 12.5, 5.0},
				{domain.Pleasures, 190.89 + 340.15, 500, 106.2, 5.31},
				{domain.Knowledge, 900.99, 500, 180.19, 9.0},
				{domain.Goals, 0, 500, 0, 0},
				{domain.FinancialInvestments, 0, 2500, 0.0, 0.0},
			},
			spent:     2108.02,
			mustSpend: 7891.98,
			used:      21.0802,
		},
		{
			name: "should handle next month with carried over excesses",
			date: nextMonth,
			entries: []expectedEntries{
				{domain.Comfort, 0, 2000, 0.0, 0.0},
				{domain.FixedCosts, 0, 4000, 0.0, 0.0},
				{domain.Pleasures, 41.04, 500, 8.2, 0.41},
				{domain.Knowledge, 400.99, 500, 80.19, 4.0},
				{domain.Goals, 0, 500, 0, 0},
				{domain.FinancialInvestments, 700.25, 2500, 28.01, 7.0},
			},
			spent:     1142.28,
			mustSpend: 8857.72,
			used:      11.4228,
		},
	}

	expenseService := NewTestExpenseService(t, tx)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)

			summary, err := expenseService.GetSummary(tt.date, user.ID)
			a.NoError(err)

			a.Equal(tt.spent, summary.Spent)
			a.Equal(tt.mustSpend, summary.MustSpend)
			a.Equal(tt.used, summary.Used)
			assertSummaryGoals(a, tt.entries, summary.Goals)
		})
	}
}
