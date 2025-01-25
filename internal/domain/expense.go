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
	Date  time.Time `json:"date"`

	GoalID uint `json:"goal_id"`
	Goal   Goal `json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SummaryEntry = struct {
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

type Summary = []SummaryEntry

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
	AllByGoalID(goalID uint, year int, month time.Month) []Expense
	GetSummary(date time.Time, goalRepo GoalRepo, salaryRepo SalaryRepo) Summary
}
