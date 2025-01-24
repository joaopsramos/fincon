package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/joaopsramos/fincon/internal/domain"
)

type ExpenseController struct {
	repo       domain.ExpenseRepository
	goalRepo   domain.GoalRepository
	salaryRepo domain.SalaryRepository
}

func NewExpenseController(repo domain.ExpenseRepository, goalRepo domain.GoalRepository, salaryRepo domain.SalaryRepository) ExpenseController {
	return ExpenseController{repo: repo, goalRepo: goalRepo, salaryRepo: salaryRepo}
}

func (c *ExpenseController) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary := c.repo.GetSummary(time.Now(), c.goalRepo, c.salaryRepo)
	json.NewEncoder(w).Encode(summary)
}
