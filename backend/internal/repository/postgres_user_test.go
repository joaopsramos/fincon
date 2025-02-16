package repository_test

import (
	"context"
	"testing"

	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/repository"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func NewTestPostgresUserRepo(t *testing.T, tx *gorm.DB) domain.UserRepo {
	t.Helper()
	return repository.NewPostgresUser(tx)
}

func TestPostgresUser_Create(t *testing.T) {
	a := assert.New(t)
	t.Parallel()
	tx := testhelper.NewTestPostgresTx(t)
	repo := NewTestPostgresUserRepo(t, tx)

	user := domain.User{Email: "test@mail.com", HashPassword: "pass"}
	salary := domain.Salary{Amount: 1000}
	a.NoError(repo.Create(context.Background(), &user, &salary))

	a.NotZero(user.ID)
	a.NotZero(salary.ID)

	goals := []map[string]any{}
	tx.Model(domain.Goal{}).Where("user_id =?", user.ID).Select("name, percentage").Scan(&goals)

	a.ElementsMatch(goals, []map[string]any{
		{"name": "Fixed costs", "percentage": int64(40)},
		{"name": "Comfort", "percentage": int64(20)},
		{"name": "Goals", "percentage": int64(5)},
		{"name": "Pleasures", "percentage": int64(5)},
		{"name": "Financial investments", "percentage": int64(25)},
		{"name": "Knowledge", "percentage": int64(5)},
	})
}
