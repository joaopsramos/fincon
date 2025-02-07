package repository_test

import (
	"testing"
	"time"

	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/repository"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func NewTestPostgresExpenseRepo(t *testing.T, tx *gorm.DB) domain.ExpenseRepo {
	t.Helper()

	return repository.NewPostgresExpense(tx)
}

func TestPostgresExpense_GetSummary(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tx := testhelper.NewTestPostgresTx(t)
	f := testhelper.NewFactory(tx)
	user := f.InsertUser()

	salaryAmount := 10_000
	f.InsertSalary(&domain.Salary{Amount: int64(salaryAmount * 100), UserID: user.ID})

	goalIDsByName := make(map[domain.GoalName]uint)

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
		goalIDsByName[goal.Name] = goal.ID
	}

	now := testhelper.MiddleOfMonth()

	expenses := []struct {
		value  float64
		date   time.Time
		goalID uint
	}{
		{50.5, now, goalIDsByName[domain.Comfort]},
		{125.49, now, goalIDsByName[domain.Comfort]},
		{400, now, goalIDsByName[domain.FixedCosts]},
		{100, now, goalIDsByName[domain.FixedCosts]},
		{190.89, now, goalIDsByName[domain.Pleasures]},
		{340.15, now, goalIDsByName[domain.Pleasures]},
		{900.99, now, goalIDsByName[domain.Knowledge]},
		{125.74, now.AddDate(0, -1, 0), goalIDsByName[domain.Comfort]},
		{500, now.AddDate(0, -1, 0), goalIDsByName[domain.Pleasures]},
		{10, now.AddDate(0, 1, 0), goalIDsByName[domain.Pleasures]},
		{700.25, now.AddDate(0, 1, 0), goalIDsByName[domain.FinancialInvestments]},
	}

	for _, e := range expenses {
		f.InsertExpense(&domain.Expense{
			Value:     int64(e.value * 100),
			Date:      e.date,
			GoalID:    e.goalID,
			CreatedAt: now,
		})
	}

	r := NewTestPostgresExpenseRepo(t, tx)
	goalRepo := NewTestPostgresGoalRepo(t, tx)
	salaryRepo := NewTestPostgresSalaryRepo(t, tx)

	type dataType struct {
		goalName  domain.GoalName
		spent     float64
		mustSpend float64
		used      float64
		total     float64
	}

	assertSummaryEntries := func(data []dataType, entriesByName map[domain.GoalName]domain.SummaryGoal) {
		for _, d := range data {
			entry := entriesByName[d.goalName]
			assert.Equal(string(d.goalName), entry.Name)
			assert.Equal(d.spent, entry.Spent.Amount)
			assert.Equal(d.mustSpend, entry.MustSpend.Amount)
			assert.Equal(d.used, float64(int(entry.Used*100))/100)
			assert.Equal(d.total, float64(int(entry.Total*100))/100)
		}
	}

	entriesByName := make(map[domain.GoalName]domain.SummaryGoal)

	for _, g := range r.GetSummary(now.AddDate(0, -1, 0), user.ID, goalRepo, salaryRepo).Goals {
		entriesByName[domain.GoalName(g.Name)] = g
	}

	data := []dataType{
		{domain.Comfort, 125.74, 2000, 6.28, 1.25},
		{domain.Comfort, 125.74, 2000, 6.28, 1.25},
		{domain.FixedCosts, 0, 4000, 0.0, 0.0},
		{domain.Pleasures, 500, 500, 100.0, 5.0},
		{domain.Knowledge, 0, 500, 0.0, 0.0},
		{domain.FinancialInvestments, 0, 2500, 0.0, 0.0},
	}

	assertSummaryEntries(data, entriesByName)

	for _, g := range r.GetSummary(now, user.ID, goalRepo, salaryRepo).Goals {
		entriesByName[domain.GoalName(g.Name)] = g
	}

	data = []dataType{
		{domain.Comfort, 50.5 + 125.49, 2000, 8.79, 1.75},
		{domain.FixedCosts, 400 + 100, 4000, 12.5, 5.0},
		{domain.Pleasures, 190.89 + 340.15, 500, 106.2, 5.31},
		{domain.Knowledge, 900.99, 500, 180.19, 9.0},
		{domain.FinancialInvestments, 0, 2500, 0.0, 0.0},
	}

	assertSummaryEntries(data, entriesByName)

	for _, g := range r.GetSummary(now.AddDate(0, 1, 0), user.ID, goalRepo, salaryRepo).Goals {
		entriesByName[domain.GoalName(g.Name)] = g
	}

	data = []dataType{
		{domain.Comfort, 0, 2000, 0.0, 0.0},
		{domain.FixedCosts, 0, 4000, 0.0, 0.0},
		{domain.Pleasures, 41.04, 500, 8.2, 0.41},
		{domain.Knowledge, 400.99, 500, 80.19, 4.0},
		{domain.FinancialInvestments, 700.25, 2500, 28.01, 7.0},
	}

	assertSummaryEntries(data, entriesByName)
}

func TestPostgresExpense_AllByGoalID(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tx := testhelper.NewTestPostgresTx(t)
	f := testhelper.NewFactory(tx)

	var user domain.User
	f.InsertUser(&user)

	goals := []domain.Goal{{Name: domain.Goals}, {Name: domain.Pleasures}, {Name: domain.Comfort}}
	for i := range len(goals) {
		f.InsertGoal(&goals[i])
	}

	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	expenses := []domain.Expense{
		{Name: "Expense 1", GoalID: goals[0].ID, Date: monthStart.AddDate(0, 0, 1), CreatedAt: monthStart},
		{Name: "Expense 2", GoalID: goals[0].ID, Date: monthStart.AddDate(0, 0, 1), CreatedAt: monthStart.Add(1 * time.Second)},
		{Name: "Expense 3", GoalID: goals[0].ID, Date: monthStart, CreatedAt: monthStart},
		{Name: "Expense 4", GoalID: goals[1].ID, Date: now, CreatedAt: now},
		{Name: "Expense 5", GoalID: goals[2].ID, Date: now, CreatedAt: now},
		{Name: "Expense 6", GoalID: goals[2].ID, Date: monthStart.AddDate(0, -1, 0), CreatedAt: now},
	}
	for i := range len(expenses) {
		expenses[i].UserID = user.ID
		f.InsertExpense(&expenses[i])
	}

	r := NewTestPostgresExpenseRepo(t, tx)
	year, month, _ := monthStart.Date()
	var actual []domain.Expense

	actual = r.AllByGoalID(goals[0].ID, year, month, user.ID)
	assert.Equal(actual[0].Name, "Expense 2")
	assert.Equal(actual[1].Name, "Expense 1")
	assert.Equal(actual[2].Name, "Expense 3")

	actual = r.AllByGoalID(goals[1].ID, year, month, user.ID)
	assert.Equal(actual[0].Name, "Expense 4")

	t.Run("filter by date", func(t *testing.T) {
		actual = r.AllByGoalID(goals[2].ID, year, month, user.ID)
		assert.Len(actual, 1)
		assert.Equal(actual[0].Name, "Expense 5")

		year, month, _ := monthStart.AddDate(0, -1, 0).Date()
		actual = r.AllByGoalID(goals[2].ID, year, month, user.ID)
		assert.Len(actual, 1)
		assert.Equal(actual[0].Name, "Expense 6")
	})
}
