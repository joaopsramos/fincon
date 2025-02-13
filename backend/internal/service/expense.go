package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/util"
)

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

type ExpenseService struct {
	expenseRepo domain.ExpenseRepo
	goalRepo    domain.GoalRepo
	salaryRepo  domain.SalaryRepo
}

func NewExpenseService(
	expenseRepo domain.ExpenseRepo,
	goalRepo domain.GoalRepo,

	salaryRepo domain.SalaryRepo,
) ExpenseService {
	return ExpenseService{expenseRepo, goalRepo, salaryRepo}
}

func (s *ExpenseService) Get(id uint, userID uuid.UUID) (*domain.Expense, error) {
	return s.expenseRepo.Get(id, userID)
}

func (s *ExpenseService) Create(dto CreateExpenseDTO, userID uuid.UUID) (*domain.Expense, error) {
	goal, err := s.goalRepo.Get(uint(dto.GoalID), userID)
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

	err = s.expenseRepo.Create(&e)

	return &e, err
}

func (s *ExpenseService) UpdateByID(id uint, dto UpdateExpenseDTO, userID uuid.UUID) (*domain.Expense, error) {
	e, err := s.expenseRepo.Get(id, userID)
	if err != nil {
		return &domain.Expense{}, err
	}

	util.UpdateIfNotZero(&e.Name, dto.Name)
	util.UpdateIfNotZero(&e.Value, int64(dto.Value*100))
	util.UpdateIfNotZero(&e.Date, dto.Date)

	err = s.expenseRepo.Update(e)

	return e, err
}

func (s *ExpenseService) Delete(id uint, userID uuid.UUID) error {
	return s.expenseRepo.Delete(id, userID)
}

func (s *ExpenseService) ChangeGoal(e *domain.Expense, goalID uint, userID uuid.UUID) error {
	goal, err := s.goalRepo.Get(goalID, userID)
	if err != nil {
		return err
	}

	e.GoalID = goal.ID

	err = s.expenseRepo.Update(e)

	return err
}

func (s *ExpenseService) AllByGoalID(goalID uint, year int, month time.Month, userID uuid.UUID) ([]domain.Expense, error) {
	return s.expenseRepo.AllByGoalID(goalID, year, month, userID)
}

func (s *ExpenseService) FindMatchingNames(name string, userID uuid.UUID) ([]string, error) {
	return s.expenseRepo.FindMatchingNames(name, userID)
}

func (s *ExpenseService) GetSummary(date time.Time, userID uuid.UUID) (domain.Summary, error) {
	salary := util.Must(s.salaryRepo.Get(userID))
	return s.expenseRepo.GetSummary(date, userID, s.goalRepo, salary)
}
