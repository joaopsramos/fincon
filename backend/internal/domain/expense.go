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

type SummaryGoal = struct {
	Name      string    `json:"name"`
	Spent     MoneyView `json:"spent"`
	MustSpend MoneyView `json:"must_spend"`
	Used      float64   `json:"used"`
	Total     float64   `json:"total"`
}

type MoneyView struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type Summary struct {
	Goals     []SummaryGoal `json:"goals"`
	Spent     MoneyView     `json:"spent"`
	MustSpend MoneyView     `json:"must_spend"`
	Used      float64       `json:"used"`
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
	Get(id uint, userID uuid.UUID) (Expense, error)
	Create(e Expense, userID uuid.UUID, goalRepo GoalRepo) (*Expense, error)
	Update(e Expense) (*Expense, error)
	Delete(id uint, userID uuid.UUID) error
	ChangeGoal(e Expense, goalID uint, userID uuid.UUID, goalRepo GoalRepo) (*Expense, error)
	AllByGoalID(goalID uint, year int, month time.Month, userID uuid.UUID) []Expense
	FindMatchingNames(name string, userID uuid.UUID) []string
	GetSummary(date time.Time, userID uuid.UUID, goalRepo GoalRepo, salaryRepo SalaryRepo) Summary
}
