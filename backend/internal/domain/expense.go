package domain

import (
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
)

type Expense struct {
	ID    uint `gorm:"primaryKey"`
	Name  string
	Value int64
	Date  time.Time `gorm:"type:timestamp without time zone"`

	UserID uuid.UUID `gorm:"type:uuid"`

	GoalID uint
	Goal   Goal

	CreatedAt time.Time
	UpdatedAt time.Time
}

type ExpenseView struct {
	ID     uint      `json:"id"`
	Name   string    `json:"name"`
	Value  MoneyView `json:"value"`
	Date   time.Time `json:"date"`
	GoalID uint      `json:"goal_id"`
}

type MonthlyGoalSpending struct {
	Goal  Goal `gorm:"embedded"`
	Date  time.Time
	Spent int64
}

type MoneyView struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

func NewMoney(money *money.Money) MoneyView {
	return MoneyView{Amount: money.AsMajorUnits(), Currency: money.Currency().Code}
}

func (e *Expense) View() ExpenseView {
	return ExpenseView{
		ID:     e.ID,
		Name:   e.Name,
		Value:  NewMoney(money.New(e.Value, money.BRL)),
		Date:   e.Date,
		GoalID: e.GoalID,
	}
}

type ExpenseRepo interface {
	Get(id uint, userID uuid.UUID) (*Expense, error)
	Create(e *Expense) error
	Update(e *Expense) error
	Delete(id uint, userID uuid.UUID) error
	AllByGoalID(goalID uint, year int, month time.Month, userID uuid.UUID) ([]Expense, error)
	FindMatchingNames(name string, userID uuid.UUID) ([]string, error)
	GetMonthlyGoalSpendings(date time.Time, userID uuid.UUID) ([]MonthlyGoalSpending, error)
}
