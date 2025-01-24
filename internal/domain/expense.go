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
	Name      string
	Spent     *money.Money
	MustSpend *money.Money
	Used      float64
	Total     float64
}

type Summary = []SummaryEntry

type ExpenseRepository interface {
	AllByGoalID(goalID uint, year int, month time.Month) []Expense
	GetSummary(date time.Time, goalRepo GoalRepository, salaryRepo SalaryRepository) Summary
}
