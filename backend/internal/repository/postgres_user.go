package repository

import (
	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/domain"
	errs "github.com/joaopsramos/fincon/internal/error"
	"gorm.io/gorm"
)

type PostgresUserRepository struct {
	db *gorm.DB
}

func NewPostgresUser(db *gorm.DB) domain.UserRepo {
	return PostgresUserRepository{db}
}

func (r PostgresUserRepository) Create(user *domain.User, salary *domain.Salary) error {
	defaultPercentages := domain.DefaulGoalPercentages()
	goals := make([]domain.Goal, 0, len(defaultPercentages))

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		salary.UserID = user.ID

		txSalaryRepo := NewPostgresSalary(tx)
		if err := txSalaryRepo.Create(salary); err != nil {
			return err
		}

		for name, percentage := range defaultPercentages {
			goals = append(goals, domain.Goal{Name: name, Percentage: percentage, UserID: user.ID})
		}

		txGoalRepo := NewPostgresGoal(tx)
		if err := txGoalRepo.Create(goals...); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r PostgresUserRepository) Get(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	result := r.db.Take(&user, id)
	return &user, result.Error
}

func (r PostgresUserRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).Take(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &domain.User{}, errs.NewNotFound("user")
		}

		return &domain.User{}, err
	}

	return &user, nil
}
