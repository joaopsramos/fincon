package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/util"
	"github.com/shopspring/decimal"
)

type ExpenseService struct {
	expenseRepo domain.ExpenseRepo
	goalRepo    domain.GoalRepo
	salaryRepo  domain.SalaryRepo
}

type CreateExpenseDTO struct {
	Name   string
	Value  float64
	Date   time.Time
	GoalID int
}

type UpdateExpenseDTO struct {
	Name  string
	Value float64
	Date  time.Time
}

type SummaryGoal = struct {
	Name      string  `json:"name"`
	Spent     float64 `json:"spent"`
	MustSpend float64 `json:"must_spend"`
	Used      float64 `json:"used"`
	Total     float64 `json:"total"`
}

type Summary struct {
	Goals     []SummaryGoal `json:"goals"`
	Spent     float64       `json:"spent"`
	MustSpend float64       `json:"must_spend"`
	Used      float64       `json:"used"`
}

func NewExpenseService(
	expenseRepo domain.ExpenseRepo,
	goalRepo domain.GoalRepo,

	salaryRepo domain.SalaryRepo,
) ExpenseService {
	return ExpenseService{expenseRepo, goalRepo, salaryRepo}
}

func (s *ExpenseService) Get(ctx context.Context, id uint, userID uuid.UUID) (*domain.Expense, error) {
	return s.expenseRepo.Get(ctx, id, userID)
}

func (s *ExpenseService) Create(ctx context.Context, dto CreateExpenseDTO, userID uuid.UUID) (*domain.Expense, error) {
	goal, err := s.goalRepo.Get(ctx, uint(dto.GoalID), userID)
	if err != nil {
		return &domain.Expense{}, err
	}

	e := domain.Expense{
		Name:   dto.Name,
		Value:  int64(dto.Value * 100),
		Date:   dto.Date,
		GoalID: goal.ID,
		UserID: userID,
	}

	err = s.expenseRepo.Create(ctx, &e)

	return &e, err
}

func (s *ExpenseService) UpdateByID(ctx context.Context, id uint, dto UpdateExpenseDTO, userID uuid.UUID) (*domain.Expense, error) {
	e, err := s.expenseRepo.Get(ctx, id, userID)
	if err != nil {
		return &domain.Expense{}, err
	}

	util.UpdateIfNotZero(&e.Name, dto.Name)
	util.UpdateIfNotZero(&e.Value, int64(dto.Value*100))
	util.UpdateIfNotZero(&e.Date, dto.Date)

	err = s.expenseRepo.Update(ctx, e)

	return e, err
}

func (s *ExpenseService) Delete(ctx context.Context, id uint, userID uuid.UUID) error {
	return s.expenseRepo.Delete(ctx, id, userID)
}

func (s *ExpenseService) ChangeGoal(ctx context.Context, e *domain.Expense, goalID uint, userID uuid.UUID) error {
	goal, err := s.goalRepo.Get(ctx, goalID, userID)
	if err != nil {
		return err
	}

	e.GoalID = goal.ID

	err = s.expenseRepo.Update(ctx, e)

	return err
}

func (s *ExpenseService) AllByGoalID(ctx context.Context, goalID uint, year int, month time.Month, userID uuid.UUID) ([]domain.Expense, error) {
	return s.expenseRepo.AllByGoalID(ctx, goalID, year, month, userID)
}

func (s *ExpenseService) FindMatchingNames(ctx context.Context, name string, userID uuid.UUID) ([]string, error) {
	return s.expenseRepo.FindMatchingNames(ctx, name, userID)
}

func (s *ExpenseService) GetSummary(ctx context.Context, date time.Time, userID uuid.UUID) (*Summary, error) {
	salary := util.Must(s.salaryRepo.Get(ctx, userID))
	monthlyGoalSpendings, err := s.expenseRepo.GetMonthlyGoalSpendings(ctx, date, userID)
	if err != nil {
		return &Summary{}, err
	}

	monthStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)

	spendingsByGoalID := make(map[uint]domain.MonthlyGoalSpending)
	for _, m := range monthlyGoalSpendings {
		date = m.Date
		goalLimit := int64(m.Goal.Percentage) * (salary.Amount / 100)

		if m.Spent <= goalLimit && date.Before(monthStart) {
			continue
		}

		if m.Spent > goalLimit {
			yearDiff := monthStart.Year() - date.Year()
			monthDiff := int(monthStart.Month()) - int(date.Month()) + yearDiff*12

			m.Spent = max(0, m.Spent-int64(monthDiff)*goalLimit)
		}

		if entry, ok := spendingsByGoalID[m.Goal.ID]; ok {
			entry.Spent += m.Spent
			spendingsByGoalID[m.Goal.ID] = entry
		} else {
			spendingsByGoalID[m.Goal.ID] = m
		}
	}

	goals := s.goalRepo.All(ctx, userID)

	var totalSpent, totalMustSpend, totalUsed decimal.Decimal

	sg := make([]SummaryGoal, len(goals))
	for i, g := range goals {
		mgs, ok := spendingsByGoalID[g.ID]
		if !ok {
			mgs = domain.MonthlyGoalSpending{}
		}

		percentage := decimal.NewFromInt(int64(g.Percentage))
		hundred := decimal.NewFromInt(100)
		spent := util.MoneyAmountToDecimal(mgs.Spent)
		salaryDec := util.MoneyAmountToDecimal(salary.Amount)

		// Calculate mustSpend (salary * percentage / 100)
		mustSpend := salaryDec.Mul(percentage).Div(hundred)

		// Calculate used percentage (100 + ((spent - mustSpend) * 100 / mustSpend))
		var used decimal.Decimal
		if !mustSpend.IsZero() {
			used = hundred.Add(spent.Sub(mustSpend).Mul(hundred).Div(mustSpend))
		}

		// Calculate total percentage (spent * 100 / salary)
		var total decimal.Decimal
		if !salaryDec.IsZero() {
			total = spent.Mul(hundred).Div(salaryDec)
		}

		sg[i] = SummaryGoal{
			Name:      string(g.Name),
			Spent:     spent.InexactFloat64(),
			MustSpend: mustSpend.InexactFloat64(),
			Used:      used.InexactFloat64(),
			Total:     total.InexactFloat64(),
		}

		totalSpent = totalSpent.Add(spent)
		totalMustSpend = salaryDec.Sub(totalSpent)
		totalUsed = totalUsed.Add(total)
	}

	return &Summary{
		Goals:     sg,
		Spent:     totalSpent.InexactFloat64(),
		MustSpend: totalMustSpend.InexactFloat64(),
		Used:      totalUsed.InexactFloat64(),
	}, nil
}
