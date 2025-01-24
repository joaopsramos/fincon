package repository

import (
	"time"

	"github.com/Rhymond/go-money"
	"github.com/joaopsramos/fincon/internal/domain"
	"gorm.io/gorm"
)

type SQLiteExpenseRepository struct {
	db *gorm.DB
}

func NewSQLiteExpense(db *gorm.DB) domain.ExpenseRepository {
	return SQLiteExpenseRepository{db}
}

func (r SQLiteExpenseRepository) AllByGoalID(goalID uint, year int, month time.Month) []domain.Expense {
	date := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	var e []domain.Expense
	r.db.
		Where("goal_id = ?", goalID).
		Where("date(date, 'start of month') = date(?, 'start of month')", date).
		Order("date DESC, created_at DESC").
		Find(&e)

	return e
}

func (r SQLiteExpenseRepository) GetSummary(date time.Time, goalRepo domain.GoalRepository, salaryRepo domain.SalaryRepository) domain.Summary {
	salary := salaryRepo.Get()

	type result struct {
		ID         uint
		Name       string
		Percentage int
		Spent      int64
		Date       string
	}

	var results []result
	r.db.Model(&domain.Goal{}).
		Joins("Expenses").
		Select("goals.id, goals.name, goals.percentage, COALESCE(date(expenses.date, 'start of month'), date('now')) date, SUM(expenses.value) spent").
		Where("date(expenses.date, 'start of month') <= date(?, 'start of month')", date).
		Group("1, 2, 3, 4").
		Scan(&results)

	monthStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)

	resultsByGoalID := make(map[uint]*result)
	for _, r := range results {
		date, _ := time.Parse(time.DateOnly, r.Date)
		goalLimit := int64(r.Percentage) * (salary.Amount / 100)

		if r.Spent <= goalLimit && date.Before(monthStart) {
			continue
		}

		if r.Spent > goalLimit {
			yearDiff := monthStart.Year() - date.Year()
			monthDiff := int(monthStart.Month()) - int(date.Month()) + yearDiff*12

			r.Spent = max(0, r.Spent-int64(monthDiff)*goalLimit)
		}

		if entry, ok := resultsByGoalID[r.ID]; ok {
			entry.Spent += r.Spent
		} else {
			resultsByGoalID[r.ID] = &r
		}

	}

	goals := goalRepo.All()

	s := make(domain.Summary, len(goals))
	for i, g := range goals {
		percentage := int64(g.Percentage)

		r, ok := resultsByGoalID[g.ID]
		if !ok {
			r = &result{}
		}

		valueSpent := money.New(r.Spent, money.BRL)
		mustSpendvalue := salary.Amount / 100 * percentage
		mustSpend := money.New(mustSpendvalue, money.BRL)

		var used float64
		mustSpendvalueF := float64(mustSpendvalue)
		used = 100 + ((float64(r.Spent) - mustSpendvalueF) * 100 / mustSpendvalueF)

		s[i] = domain.SummaryEntry{
			Name:      string(g.Name),
			Spent:     valueSpent,
			MustSpend: mustSpend,
			Used:      used,
			Total:     float64(r.Spent*100) / float64(salary.Amount),
		}
	}

	return s
}
