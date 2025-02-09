package repository_test

import (
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
	assert := assert.New(t)
	t.Parallel()
	tx := testhelper.NewTestPostgresTx(t)
	r := NewTestPostgresUserRepo(t, tx)

	user := domain.User{Email: "test@mail.com", HashPassword: "pass"}
	assert.NoError(r.Create(&user))

	assert.NotZero(user.ID)

	goals := []map[string]any{}
	tx.Model(domain.Goal{}).Where("user_id =?", user.ID).Select("name, percentage").Scan(&goals)

	assert.ElementsMatch(goals, []map[string]any{
		{"name": "Fixed costs", "percentage": int64(40)},
		{"name": "Comfort", "percentage": int64(20)},
		{"name": "Goals", "percentage": int64(5)},
		{"name": "Pleasures", "percentage": int64(5)},
		{"name": "Financial investments", "percentage": int64(25)},
		{"name": "Knowledge", "percentage": int64(5)},
	})
}
