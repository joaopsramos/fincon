package domain

import (
	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
)

type Salary struct {
	ID     uint `gorm:"primaryKey"`
	Amount int64

	UserID uuid.UUID `gorm:"type:uuid"`
}

type SalaryView struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

func (s *Salary) View() SalaryView {
	value := NewMoney(money.New(s.Amount, money.BRL))
	return SalaryView(value)
}

type SalaryRepo interface {
	Get(userID uuid.UUID) (*Salary, error)
	Create(salary *Salary) error
	Update(salary *Salary) error
}
