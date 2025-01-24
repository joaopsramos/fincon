package domain

import (
	"time"

	"github.com/Rhymond/go-money"
)

type Expense struct {
	ID    uint `gorm:"primaryKey"`
	Name  string
	Value int64
	Date  time.Time

	GoalID uint
	Goal   Goal

	CreatedAt time.Time
	UpdatedAt time.Time
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

type ExpenseRepository interface {
	AllByGoalID(goalID uint, year int, month time.Month) []Expense
	GetSummary(date time.Time, goalRepo GoalRepository, salaryRepo SalaryRepository) Summary
}
