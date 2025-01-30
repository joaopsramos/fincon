package domain

import (
	"encoding/json"
	"time"

	"github.com/Rhymond/go-money"
)

type Expense struct {
	ID    uint      `json:"id" gorm:"primaryKey"`
	Name  string    `json:"name"`
	Value int64     `json:"value"`
	Date  time.Time `gorm:"type:timestamp without time zone" json:"date"`

	GoalID uint `json:"goal_id"`
	Goal   Goal `json:"-"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type SummaryGoal = struct {
	Name      string  `json:"name"`
	Spent     Money   `json:"spent"`
	MustSpend Money   `json:"must_spend"`
	Used      float64 `json:"used"`
	Total     float64 `json:"total"`
}

type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type Summary struct {
	Goals     []SummaryGoal `json:"goals"`
	Spent     Money         `json:"spent"`
	MustSpend Money         `json:"must_spend"`
	Used      float64       `json:"used"`
}

func NewMoney(money *money.Money) Money {
	return Money{Amount: money.AsMajorUnits(), Currency: money.Currency().Code}
}

func (e *Expense) MarshalJSON() ([]byte, error) {
	type Alias Expense

	return json.Marshal(&struct {
		Value Money `json:"value"`
		*Alias
	}{
		Value: NewMoney(money.New(e.Value, money.BRL)),
		Alias: (*Alias)(e),
	})
}

type ExpenseRepo interface {
	Get(id uint) (*Expense, error)
	Create(e Expense, goalRepo GoalRepo) (*Expense, error)
	Update(e Expense) (*Expense, error)
	Delete(id uint) error
	ChangeGoal(e Expense, goalID uint) (*Expense, error)
	AllByGoalID(goalID uint, year int, month time.Month) []Expense
	GetSummary(date time.Time, goalRepo GoalRepo, salaryRepo SalaryRepo) Summary
}
