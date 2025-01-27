package testhelper

import (
	"github.com/joaopsramos/fincon/internal/domain"
	"gorm.io/gorm"
)

type Factory struct {
	tx *gorm.DB
}

func NewFactory(tx *gorm.DB) *Factory {
	return &Factory{tx: tx}
}

func (f *Factory) InsertSalary(s *domain.Salary) {
	f.tx.Create(s)
}

func (f *Factory) InsertGoal(g *domain.Goal) {
	f.tx.Create(g)
}

func (f *Factory) InsertExpense(e *domain.Expense) {
	f.tx.Create(e)
}
