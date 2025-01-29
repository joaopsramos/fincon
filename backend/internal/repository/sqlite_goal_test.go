package repository_test

import (
	"testing"

	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/repository"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func NewTestSQLiteGoalRepo(t *testing.T, tx *gorm.DB) domain.GoalRepo {
	return repository.NewSQLiteGoal(tx)
}

func TestSQLiteGoal_All(t *testing.T) {
	t.Parallel()

	tx := testhelper.NewTestSQLiteTx(t)
	f := testhelper.NewFactory(tx)

	goals := []domain.Goal{{Name: domain.Goals, Percentage: 20}, {Name: domain.Pleasures, Percentage: 80}}
	for i := range len(goals) {
		f.InsertGoal(&goals[i])
	}

	r := NewTestSQLiteGoalRepo(t, tx)
	assert.Equal(t, goals, r.All())
}
