package repository_test

import (
	"testing"

	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/repository"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func NewTestPostgresSalaryRepo(t *testing.T, tx *gorm.DB) domain.SalaryRepo {
	return repository.NewPostgresSalary(tx)
}

func TestPostgresSalary_Get(t *testing.T) {
	t.Parallel()

	tx := testhelper.NewTestPostgresTx(t)

	f := testhelper.NewFactory(tx)
	f.InsertSalary(&domain.Salary{Amount: 200})
	r := NewTestPostgresSalaryRepo(t, tx)

	assert.Equal(t, int64(200), r.Get().Amount)
}
