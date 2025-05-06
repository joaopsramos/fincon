package repository_test

import (
	"context"
	"slices"
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

func TestPostgresExpense_GetMonthlyGoalSpendings(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tx := testhelper.NewTestPostgresTx(t)
	f := testhelper.NewFactory(tx)
	user1 := f.InsertUser()
	user2 := f.InsertUser()
	users := []domain.User{user1, user2}

	// only of user 1
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

	var goals []domain.Goal
	for _, u := range users {
		for _, g := range goalsToInsert {
			goal := domain.Goal{Name: g.name, Percentage: g.percentage, UserID: u.ID}
			f.InsertGoal(&goal)

			if u.ID == user1.ID {
				goalsByName[goal.Name] = goal
			}

			goals = append(goals, goal)
		}
	}

	now := testhelper.MiddleOfMonth()

	expenses := []struct {
		value    float64
		date     time.Time
		goalName domain.GoalName
	}{
		{50.5, now, domain.Comfort},
		{125.49, now, domain.Comfort},
		{400, now, domain.FixedCosts},
		{100, now, domain.FixedCosts},
		{190.89, now, domain.Pleasures},
		{340.15, now, domain.Pleasures},
		{900.99, now, domain.Knowledge},
		{125.74, now.AddDate(0, -1, 0), domain.Comfort},
		{500, now.AddDate(0, -1, 0), domain.Pleasures},
		{10, now.AddDate(0, 1, 0), domain.Pleasures},
		{700.25, now.AddDate(0, 1, 0), domain.FinancialInvestments},
	}

	for _, u := range users {
		for _, e := range expenses {
			goalIdx := slices.IndexFunc(goals, func(g domain.Goal) bool {
				return g.UserID == u.ID && g.Name == e.goalName
			})
			goal := goals[goalIdx]

			f.InsertExpense(&domain.Expense{
				Value:     int64(e.value * 100),
				Date:      e.date,
				GoalID:    goal.ID,
				UserID:    u.ID,
				CreatedAt: now,
			})
		}
	}

	type expectedSpending struct {
		goal  domain.Goal
		spent int
		date  time.Time
	}

	assertSpendings := func(tests []expectedSpending, spendings []domain.MonthlyGoalSpending) {
		for _, tt := range tests {
			monthStart := time.Date(tt.date.Year(), tt.date.Month(), 1, 0, 0, 0, 0, time.UTC)
			idx := slices.IndexFunc(spendings, func(s domain.MonthlyGoalSpending) bool {
				return s.Goal.Name == tt.goal.Name && s.Date.Equal(monthStart)
			})

			spending := spendings[idx]
			assert.Equal(tt.goal.Name, spending.Goal.Name)
			assert.Equal(tt.goal.ID, spending.Goal.ID)
			assert.Equal(tt.goal.Percentage, spending.Goal.Percentage)
			assert.Equal(int64(tt.spent), spending.Spent)
			assert.Equal(monthStart, spending.Date)
		}

		assert.Equal(len(tests), len(spendings))
	}

	oneMonthAgo := now.AddDate(0, -1, 0)
	twoMonthsAgo := now.AddDate(0, -2, 0)
	nextMonth := now.AddDate(0, 1, 0)

	tests := []struct {
		name      string
		queryDate time.Time
		expected  []expectedSpending
	}{
		{
			"empty result when no expenses exist",
			twoMonthsAgo,
			[]expectedSpending{},
		},
		{
			"previous month spendings",
			oneMonthAgo,
			[]expectedSpending{
				{goalsByName[domain.Comfort], 12574, oneMonthAgo},
				{goalsByName[domain.Pleasures], 50000, oneMonthAgo},
			},
		},
		{
			"current month includes previous month's spendings",
			now,
			[]expectedSpending{
				{goalsByName[domain.Comfort], 12574, oneMonthAgo},
				{goalsByName[domain.Comfort], 17599, now},
				{goalsByName[domain.FixedCosts], 50000, now},
				{goalsByName[domain.Pleasures], 50000, oneMonthAgo},
				{goalsByName[domain.Pleasures], 53104, now},
				{goalsByName[domain.Knowledge], 90099, now},
			},
		},
		{
			"next month includes all previous spendings",
			nextMonth,
			[]expectedSpending{
				{goalsByName[domain.Comfort], 12574, oneMonthAgo},
				{goalsByName[domain.Comfort], 17599, now},
				{goalsByName[domain.FixedCosts], 50000, now},
				{goalsByName[domain.Pleasures], 50000, oneMonthAgo},
				{goalsByName[domain.Pleasures], 53104, now},
				{goalsByName[domain.Knowledge], 90099, now},
				{goalsByName[domain.FinancialInvestments], 70025, nextMonth},
				{goalsByName[domain.Pleasures], 1000, nextMonth},
			},
		},
	}

	repo := NewTestPostgresExpenseRepo(t, tx)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spendings, err := repo.GetMonthlyGoalSpendings(context.Background(), tt.queryDate, user1.ID)
			assert.NoError(err)
			assertSpendings(tt.expected, spendings)
		})
	}
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
		goals[i].UserID = user.ID
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

	actual, err := r.AllByGoalID(context.Background(), goals[0].ID, year, month, user.ID)
	assert.NoError(err)
	assert.Equal(actual[0].Name, "Expense 2")
	assert.Equal(actual[1].Name, "Expense 1")
	assert.Equal(actual[2].Name, "Expense 3")

	actual, err = r.AllByGoalID(context.Background(), goals[1].ID, year, month, user.ID)
	assert.NoError(err)
	assert.Equal(actual[0].Name, "Expense 4")

	t.Run("filter by date", func(t *testing.T) {
		actual, err := r.AllByGoalID(context.Background(), goals[2].ID, year, month, user.ID)
		assert.NoError(err)
		assert.Len(actual, 1)
		assert.Equal(actual[0].Name, "Expense 5")

		year, month, _ := monthStart.AddDate(0, -1, 0).Date()
		actual, err = r.AllByGoalID(context.Background(), goals[2].ID, year, month, user.ID)
		assert.NoError(err)
		assert.Len(actual, 1)
		assert.Equal(actual[0].Name, "Expense 6")
	})
}
