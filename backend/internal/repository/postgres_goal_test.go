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

func NewTestPostgresGoalRepo(t *testing.T, tx *gorm.DB) domain.GoalRepo {
	return repository.NewPostgresGoal(tx)
}

func TestPostgresGoal_All(t *testing.T) {
	t.Parallel()

	tx := testhelper.NewTestPostgresTx(t)
	f := testhelper.NewFactory(tx)

	user := f.InsertUser()
	goals := []domain.Goal{
		{Name: domain.Goals, Percentage: 20, UserID: user.ID},
		{Name: domain.Pleasures, Percentage: 80, UserID: user.ID},
	}
	for i := range goals {
		f.InsertGoal(&goals[i])
	}

	r := NewTestPostgresGoalRepo(t, tx)
	assert.Equal(t, goals, r.All(context.Background(), user.ID))
}
