package repository_test

import (
	"testing"

	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/repository"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/stretchr/testify/assert"
)

func NewTestSQLiteSalaryRepo(t *testing.T) domain.SalaryRepository {
	db := testhelper.NewTestSQLiteDB()

	return repository.NewSQLiteSalary(db)
}

func TestSQLiteSalary_Get(t *testing.T) {
	r := NewTestSQLiteSalaryRepo(t)
	assert.Equal(t, r.Get().Amount, int64(0))
	assert.Equal(t, 2, 2)
}
