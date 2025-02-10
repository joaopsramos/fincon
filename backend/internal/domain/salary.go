package domain

import "github.com/google/uuid"

type Salary struct {
	Amount int64

	UserID uuid.UUID `gorm:"type:uuid"`
}

type SalaryView struct {
	Amount int64 `json:"amount"`
}

func (s *Salary) View() SalaryView {
	return SalaryView{
		Amount: s.Amount,
	}
}

type SalaryRepo interface {
	Get(userID uuid.UUID) Salary
}
