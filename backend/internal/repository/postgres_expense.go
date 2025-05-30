package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/errs"
	"gorm.io/gorm"
)

type PostgresExpenseRepository struct {
	db *gorm.DB
}

func NewPostgresExpense(db *gorm.DB) domain.ExpenseRepo {
	return PostgresExpenseRepository{db}
}

func (r PostgresExpenseRepository) FindMatchingNames(ctx context.Context, name string, userID uuid.UUID) ([]string, error) {
	var names []string
	result := r.db.
		WithContext(ctx).
		Model(&domain.Expense{}).
		Where("user_id = ?", userID).
		Where("unaccent(name) ILIKE unaccent(?)", "%"+name+"%").
		Distinct("name").Pluck("name", &names)

	return names, result.Error
}

func (r PostgresExpenseRepository) Get(ctx context.Context, id uint, userID uuid.UUID) (*domain.Expense, error) {
	var e domain.Expense

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Take(&e, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &domain.Expense{}, errs.NewNotFound("expense")
		}

		return &domain.Expense{}, err
	}

	return &e, nil
}

func (r PostgresExpenseRepository) Create(ctx context.Context, e *domain.Expense) error {
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return err
	}

	return nil
}

func (r PostgresExpenseRepository) CreateMany(ctx context.Context, e []domain.Expense) error {
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return err
	}

	return nil
}

func (r PostgresExpenseRepository) Update(ctx context.Context, e *domain.Expense) error {
	if err := r.db.WithContext(ctx).Model(e).Updates(e).Error; err != nil {
		return err
	}

	return nil
}

func (r PostgresExpenseRepository) Delete(ctx context.Context, id uint, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&domain.Expense{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errs.NewNotFound("expense")
	}

	return nil
}

func (r PostgresExpenseRepository) AllByGoalID(ctx context.Context, goalID uint, year int, month time.Month, userID uuid.UUID) ([]domain.Expense, error) {
	date := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	var e []domain.Expense
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("goal_id = ?", goalID).
		Where("date_trunc('month', date) = date_trunc('month', ?::timestamp)", date).
		Order("date DESC, created_at DESC").
		Find(&e)

	return e, result.Error
}

func (r PostgresExpenseRepository) GetMonthlyGoalSpendings(ctx context.Context, date time.Time, userID uuid.UUID) ([]domain.MonthlyGoalSpending, error) {
	var monthlyGoalSpendings []domain.MonthlyGoalSpending
	err := r.db.WithContext(ctx).Model(&domain.Goal{}).
		Joins("JOIN expenses ON goals.id = expenses.goal_id").
		Where("date_trunc('month', expenses.date) <= date_trunc('month', ?::date)", date).
		Where("goals.user_id = ?", userID).
		Select("goals.*, date_trunc('month', expenses.date) date, SUM(expenses.value) spent").
		Group("goals.id, date_trunc('month', expenses.date)").
		Scan(&monthlyGoalSpendings).Error
	if err != nil {
		return []domain.MonthlyGoalSpending{}, err
	}

	return monthlyGoalSpendings, nil
}
