package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/errs"
	"gorm.io/gorm"
)

type PostgresUserRepository struct {
	db *gorm.DB
}

func NewPostgresUser(db *gorm.DB) domain.UserRepo {
	return PostgresUserRepository{db}
}

func (r PostgresUserRepository) Create(ctx context.Context, user *domain.User, salary *domain.Salary) error {
	defaultPercentages := domain.DefaultGoalPercentages()
	goals := make([]domain.Goal, 0, len(defaultPercentages))

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		salary.UserID = user.ID

		txSalaryRepo := NewPostgresSalary(tx)
		if err := txSalaryRepo.Create(context.Background(), salary); err != nil {
			return err
		}

		for name, percentage := range defaultPercentages {
			goals = append(goals, domain.Goal{Name: name, Percentage: percentage, UserID: user.ID})
		}

		txGoalRepo := NewPostgresGoal(tx)
		if err := txGoalRepo.Create(context.Background(), goals...); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (r PostgresUserRepository) Get(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	result := r.db.Take(&user, id)
	return &user, result.Error
}

func (r PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).Take(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.NewNotFound("user")
		}

		return nil, err
	}

	return &user, nil
}

func (r PostgresUserRepository) CreateToken(ctx context.Context, token *domain.UserToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r PostgresUserRepository) GetUserTokenByToken(ctx context.Context, token string) (*domain.UserToken, error) {
	var userToken domain.UserToken
	err := r.db.Where("token = ?", token).Take(&userToken).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.NewNotFound("token")
		}

		return nil, err
	}

	return &userToken, nil
}

func (r PostgresUserRepository) UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	result := r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", userID).Update("hash_password", hashedPassword)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errs.NewNotFound("user")
	}

	return nil
}

func (r PostgresUserRepository) MarkTokenAsUsed(ctx context.Context, tokenID uint) error {
	result := r.db.WithContext(ctx).Model(&domain.UserToken{}).Where("id = ?", tokenID).Update("used", true)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errs.NewNotFound("token")
	}

	return nil
}
