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
	encoder := json.NewEncoder(w)
	query := r.URL.Query()
	date := time.Now()

	if queryDate := query.Get("date"); queryDate != "" {
		parsedDate, err := time.Parse("2006-01-02", queryDate)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(map[string]any{"error": "invalid date"})
			return
		}

		date = parsedDate
	}

	summary := c.repo.GetSummary(date, c.goalRepo, c.salaryRepo)
	encoder.Encode(summary)
}
