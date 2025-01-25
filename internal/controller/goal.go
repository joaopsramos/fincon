package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joaopsramos/fincon/internal/domain"
)

type GoalController struct {
	repo        domain.GoalRepo
	expenseRepo domain.ExpenseRepo
}

func NewGoalController(repo domain.GoalRepo, expenseRepo domain.ExpenseRepo) GoalController {
	return GoalController{repo: repo, expenseRepo: expenseRepo}
}

func (c *GoalController) Index(w http.ResponseWriter, r *http.Request) {
	goals := c.repo.All()
	json.NewEncoder(w).Encode(goals)
}

func (c *GoalController) GetExpenses(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	query := r.URL.Query()
	now := time.Now()
	year, month, _ := now.Date()

	queryYear := query.Get("year")
	queryMonth := query.Get("month")

	if queryYear != "" {
		parsedYear, err := strconv.Atoi(queryYear)
		if err != nil || parsedYear < 1 {
			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(map[string]any{"error": "invalid year"})
			return
		}

		year = parsedYear
	}

	if queryMonth != "" {
		parsedMonth, err := strconv.Atoi(queryMonth)
		if err != nil || parsedMonth < 1 || parsedMonth > 12 {
			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(map[string]any{"error": "invalid month"})
			return
		}

		month = time.Month(parsedMonth)
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(map[string]any{"error": "invalid goal id"})
		return
	}

	expenses := c.expenseRepo.AllByGoalID(uint(id), year, month)
	encoder.Encode(expenses)
}
