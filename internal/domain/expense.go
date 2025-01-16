package domain

import "time"

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

type ExpenseRepository interface {
	GetByGoalID(goalID uint, conditions ...any) []Expense
}
