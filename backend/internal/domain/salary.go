package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/util"
)

type Salary struct {
	ID     uint `gorm:"primaryKey"`
	Amount int64

	UserID uuid.UUID `gorm:"type:uuid"`
}

type SalaryDTO struct {
	Amount float64 `json:"amount"`
}

func (s *Salary) ToDTO() SalaryDTO {
	return SalaryDTO{Amount: util.MoneyAmountToFloat(s.Amount)}
}

type SalaryRepo interface {
	Get(ctx context.Context, userID uuid.UUID) (*Salary, error)
	Create(ctx context.Context, salary *Salary) error
	Update(ctx context.Context, salary *Salary) error
}
