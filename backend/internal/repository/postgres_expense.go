package repository

import (
	"errors"
	"log/slog"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/domain"
	"gorm.io/gorm"
)

type PostgresExpenseRepository struct {
	db *gorm.DB
}

func NewPostgresExpense(db *gorm.DB) domain.ExpenseRepo {
	return PostgresExpenseRepository{db}
}

func (r PostgresExpenseRepository) FindMatchingNames(name string, userID uuid.UUID) []string {
	var names []string
	r.db.Model(&domain.Expense{}).Where("user_id = ?", userID).Where("unaccent(name) ILIKE unaccent(?)", "%"+name+"%").Distinct("name").Pluck("name", &names)

	return names
}

func (r PostgresExpenseRepository) Get(id uint, userID uuid.UUID) (domain.Expense, error) {
	var e domain.Expense
	result := r.db.Where("user_id = ?", userID).Take(&e, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return domain.Expense{}, errors.New("expense not found")
		}

		slog.Error(result.Error.Error())
		return domain.Expense{}, errors.New("expense could not be retrieved")
	}

	return e, nil
}

func (r PostgresExpenseRepository) Create(e domain.Expense, userID uuid.UUID, goalRepo domain.GoalRepo) (*domain.Expense, error) {
	_, err := goalRepo.Get(e.GoalID, userID)
	if err != nil {
		return &domain.Expense{}, err
	}

	e.UserID = userID

	result := r.db.Create(&e)
	if result.Error != nil {
		slog.Error(result.Error.Error())
		return &domain.Expense{}, errors.New("expense could not be created")
	}

	return &e, nil
}

func (r PostgresExpenseRepository) Update(e domain.Expense) (*domain.Expense, error) {
	result := r.db.Model(&e).Select("Name", "Value", "Date").Updates(e)
	if result.Error != nil {
		slog.Error(result.Error.Error())
		return &domain.Expense{}, errors.New("expense could not be updated")
	}

	return &e, nil
}

func (r PostgresExpenseRepository) ChangeGoal(
	e domain.Expense,
	goalID uint,
	userID uuid.UUID,
	goalRepo domain.GoalRepo,
) (*domain.Expense, error) {
	goal, err := goalRepo.Get(goalID, userID)
	if err != nil {
		return &domain.Expense{}, err
	}

	result := r.db.Model(&e).Update("goal_id", goal.ID)
	if result.Error != nil {
		slog.Error(result.Error.Error())
		return &domain.Expense{}, errors.New("expense goal could not be changed")
	}

	return &e, nil
}

func (r PostgresExpenseRepository) Delete(id uint, userID uuid.UUID) error {
	result := r.db.Where("user_id = ?", userID).Delete(&domain.Expense{}, id)
	if result.Error != nil {
		slog.Error(result.Error.Error())
		return errors.New("expense could not be deleted")
	}

	if result.RowsAffected == 0 {
		return errors.New("expense not found")
	}

	return nil
}

func (r PostgresExpenseRepository) AllByGoalID(goalID uint, year int, month time.Month, userID uuid.UUID) []domain.Expense {
	date := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	var e []domain.Expense
	r.db.
		Where("user_id = ?", userID).
		Where("goal_id = ?", goalID).
		Where("date_trunc('month', date) = date_trunc('month', ?::timestamp)", date).
		Order("date DESC, created_at DESC").
		Find(&e)

	return e
}

func (r PostgresExpenseRepository) GetSummary(date time.Time, userID uuid.UUID, goalRepo domain.GoalRepo, salaryRepo domain.SalaryRepo) domain.Summary {
	salary := salaryRepo.Get(userID)

	type result struct {
		ID         uint
		Name       string
		Percentage int
		Spent      int64
		Date       time.Time
	}

	var results []result
	r.db.Model(&domain.Goal{}).
		Joins("JOIN expenses ON goals.id = expenses.goal_id").
		Select("goals.id, goals.name, goals.percentage, date_trunc('month', expenses.date) date, SUM(expenses.value) spent").
		Where("date_trunc('month', expenses.date) <= date_trunc('month', ?::date)", date).
		Where("goals.user_id = ?", userID).
		Group("1, 2, 3, 4").
		Scan(&results)

	monthStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)

	resultsByGoalID := make(map[uint]*result)
	for _, r := range results {
		date = r.Date
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

	goals := goalRepo.All(userID)

	totalSpent := money.New(0, money.BRL)
	totalMustSpend := money.New(0, money.BRL)
	totalUsed := 0.0

	sg := make([]domain.SummaryGoal, len(goals))
	for i, g := range goals {
		percentage := int64(g.Percentage)

		r, ok := resultsByGoalID[g.ID]
		if !ok {
			r = &result{}
		}

		valueSpent := money.New(r.Spent, money.BRL)
		mustSpendvalue := salary.Amount / 100 * percentage
		mustSpend := money.New(mustSpendvalue, money.BRL)

		mustSpendvalueF := float64(mustSpendvalue)
		used := 100 + ((float64(r.Spent) - mustSpendvalueF) * 100 / mustSpendvalueF)
		total := float64(r.Spent*100) / float64(salary.Amount)

		sg[i] = domain.SummaryGoal{
			Name:      string(g.Name),
			Spent:     domain.NewMoney(valueSpent),
			MustSpend: domain.NewMoney(mustSpend),
			Used:      used,
			Total:     total,
		}

		totalSpent, _ = totalSpent.Add(valueSpent)
		totalMustSpend = money.New(salary.Amount-totalSpent.Amount(), money.BRL)
		totalUsed += total
	}

	return domain.Summary{
		Goals:     sg,
		Spent:     domain.NewMoney(totalSpent),
		MustSpend: domain.NewMoney(totalMustSpend),
		Used:      totalUsed,
	}
}
