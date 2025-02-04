package domain

import "github.com/google/uuid"

type Salary struct {
	Amount int64 `json:"amount"`

	UserID uuid.UUID `json:"-" gorm:"type:uuid"`
}

type SalaryRepo interface {
	Get(userID uuid.UUID) Salary
}
