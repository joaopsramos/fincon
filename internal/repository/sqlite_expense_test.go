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

func NewTestSQLiteExpenseRepo(t *testing.T, tx *gorm.DB) domain.ExpenseRepo {
	t.Helper()

	return repository.NewSQLiteExpense(tx)
}

func TestSQLiteExpense_GetSummary(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tx := testhelper.NewTestSQLiteDB().Begin()
	t.Cleanup(func() {
		tx.Rollback()
	})

	f := testhelper.NewFactory(tx)

	salaryAmount := 10_000
	f.InsertSalary(&domain.Salary{Amount: int64(salaryAmount * 100)})

	goalIDsByName := make(map[domain.GoalName]uint)

	goalsToInsert := []struct {
		name       domain.GoalName
		percentage uint
	}{
		{domain.Comfort, 20},
		{domain.FixedCosts, 40},
		{domain.Goals, 5},
		{domain.Pleasures, 5},
		{domain.FinancialInvestments, 25},
		{domain.Knowledge, 5},
	}
	for _, g := range goalsToInsert {
		goal := domain.Goal{Name: g.name, Percentage: g.percentage}
		f.InsertGoal(&goal)
		goalIDsByName[goal.Name] = goal.ID
	}

	now := time.Now().UTC()

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

	r := NewTestSQLiteExpenseRepo(t, tx)
	goalRepo := NewTestSQLiteGoalRepo(t, tx)
	salaryRepo := NewTestSQLiteSalaryRepo(t, tx)

	entriesByName := make(map[domain.GoalName]domain.SummaryEntry)

	for _, e := range r.GetSummary(now.AddDate(0, -1, 0), goalRepo, salaryRepo) {
		entriesByName[domain.GoalName(e.Name)] = e
	}

	assertSummaryEntry(domain.Comfort, 125.74, 2000, 6.28, 1.25, entriesByName, assert)
	assertSummaryEntry(domain.FixedCosts, 0, 4000, 0.0, 0.0, entriesByName, assert)
	assertSummaryEntry(domain.Pleasures, 500, 500, 100.0, 5.0, entriesByName, assert)
	assertSummaryEntry(domain.Knowledge, 0, 500, 0.0, 0.0, entriesByName, assert)
	assertSummaryEntry(domain.FinancialInvestments, 0, 2500, 0.0, 0.0, entriesByName, assert)

	for _, e := range r.GetSummary(now, goalRepo, salaryRepo) {
		entriesByName[domain.GoalName(e.Name)] = e
	}

	assertSummaryEntry(domain.Comfort, 50.5+125.49, 2000, 8.79, 1.75, entriesByName, assert)
	assertSummaryEntry(domain.FixedCosts, 400+100, 4000, 12.5, 5.0, entriesByName, assert)
	assertSummaryEntry(domain.Pleasures, 190.89+340.15, 500, 106.2, 5.31, entriesByName, assert)
	assertSummaryEntry(domain.Knowledge, 900.99, 500, 180.19, 9.0, entriesByName, assert)
	assertSummaryEntry(domain.FinancialInvestments, 0, 2500, 0.0, 0.0, entriesByName, assert)

	for _, e := range r.GetSummary(now.AddDate(0, 1, 0), goalRepo, salaryRepo) {
		entriesByName[domain.GoalName(e.Name)] = e
	}

	assertSummaryEntry(domain.Comfort, 0, 2000, 0.0, 0.0, entriesByName, assert)
	assertSummaryEntry(domain.FixedCosts, 0, 4000, 0.0, 0.0, entriesByName, assert)
	assertSummaryEntry(domain.Pleasures, 31.04, 500, 6.2, 0.31, entriesByName, assert)
	assertSummaryEntry(domain.Knowledge, 400.99, 500, 80.19, 4.0, entriesByName, assert)
	assertSummaryEntry(domain.FinancialInvestments, 700.25, 2500, 28.01, 7.0, entriesByName, assert)
}

func TestSQLiteExpense_GetByGoalID(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tx := testhelper.NewTestSQLiteDB().Begin()
	t.Cleanup(func() {
		tx.Rollback()
	})

	f := testhelper.NewFactory(tx)

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
		f.InsertExpense(&expenses[i])
	}

	r := NewTestSQLiteExpenseRepo(t, tx)
	year, month, _ := monthStart.Date()
	var actual []domain.Expense

	actual = r.AllByGoalID(goals[0].ID, year, month)
	assert.Equal(actual[0].Name, "Expense 2")
	assert.Equal(actual[1].Name, "Expense 1")
	assert.Equal(actual[2].Name, "Expense 3")

	actual = r.AllByGoalID(goals[1].ID, year, month)
	assert.Equal(actual[0].Name, "Expense 4")

	t.Run("filter by date", func(t *testing.T) {
		actual = r.AllByGoalID(goals[2].ID, year, month)
		assert.Len(actual, 1)
		assert.Equal(actual[0].Name, "Expense 5")

		year, month, _ := monthStart.AddDate(0, -1, 0).Date()
		actual = r.AllByGoalID(goals[2].ID, year, month)
		assert.Len(actual, 1)
		assert.Equal(actual[0].Name, "Expense 6")
	})
}

func assertSummaryEntry(
	name domain.GoalName,
	spent float64,
	mustSpend float64,
	used float64,
	total float64,
	entriesByName map[domain.GoalName]domain.SummaryEntry,
	assert *assert.Assertions,
) {
	entry := entriesByName[name]

	assert.Equal(string(name), entry.Name)
	assert.Equal(spent, entry.Spent.Amount)
	assert.Equal(mustSpend, entry.MustSpend.Amount)
	assert.Equal(used, float64(int(entry.Used*100))/100)
	assert.Equal(total, float64(int(entry.Total*100))/100)
}
