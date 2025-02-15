package domain

import (
	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/util"
)

type Salary struct {
	ID     uint `gorm:"primaryKey"`
	Amount int64

	UserID uuid.UUID `gorm:"type:uuid"`
}

type SalaryView struct {
	Amount float64 `json:"amount"`
}

func (s *Salary) View() SalaryView {
	return SalaryView{Amount: util.MoneyAmountToFloat(s.Amount)}
}

type SalaryRepo interface {
	Get(userID uuid.UUID) (*Salary, error)
	Create(salary *Salary) error
	Update(salary *Salary) error
}
