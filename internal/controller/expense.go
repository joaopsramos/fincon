package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joaopsramos/fincon/internal/domain"
	"gorm.io/gorm"
)

type ExpenseController struct {
	expenseRepo domain.ExpenseRepo
	goalRepo    domain.GoalRepo
	salaryRepo  domain.SalaryRepo
}

func NewExpenseController(
	expenseRepo domain.ExpenseRepo,
	goalRepo domain.GoalRepo,
	salaryRepo domain.SalaryRepo,
) ExpenseController {
	return ExpenseController{expenseRepo: expenseRepo, goalRepo: goalRepo, salaryRepo: salaryRepo}
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

	summary := c.expenseRepo.GetSummary(date, c.goalRepo, c.salaryRepo)
	encoder.Encode(summary)
}

func (c *ExpenseController) Create(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)

	var params struct {
		Name   string  `json:"name"`
		Value  float64 `json:"value"`
		Date   string  `json:"date"`
		GoalID uint    `json:"goal_id"`
	}
	json.NewDecoder(r.Body).Decode(&params)

	date, err := time.Parse("02/01/2006", params.Date)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(map[string]any{"error": "invalid date"})
		return
	}

	toCreate := domain.Expense{
		Name:   params.Name,
		Value:  int64(params.Value * 100),
		Date:   date,
		GoalID: params.GoalID,
	}

	expense, err := c.expenseRepo.Create(toCreate, c.goalRepo)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(map[string]any{"error": "goal not found"})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(map[string]any{"error": "internal server error"})
		return
	}

	encoder.Encode(expense)
}

func (c *ExpenseController) Update(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)

	var params struct {
		Name  string  `json:"name"`
		Value float64 `json:"value"`
		Date  string  `json:"date"`
	}
	json.NewDecoder(r.Body).Decode(&params)

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(map[string]any{"error": "invalid expense id"})
		return
	}

	date, err := time.Parse("02/01/2006", params.Date)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(map[string]any{"error": "invalid date"})
		return
	}

	toUpdate, err := c.expenseRepo.Get(uint(id))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(map[string]any{"error": "expense not found"})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(map[string]any{"error": "internal server error"})
		return
	}

	toUpdate.Name = params.Name
	toUpdate.Value = int64(params.Value * 100)
	toUpdate.Date = date

	expense, err := c.expenseRepo.Update(*toUpdate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(map[string]any{"error": "internal server error"})
		return
	}

	encoder.Encode(expense)
}

func (c *ExpenseController) UpdateGoal(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)

	var params struct {
		GoalID uint `json:"goal_id"`
	}
	json.NewDecoder(r.Body).Decode(&params)

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(map[string]any{"error": "invalid expense id"})
		return
	}

	toUpdate, err := c.expenseRepo.Get(uint(id))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(map[string]any{"error": "expense not found"})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(map[string]any{"error": "internal server error"})
		return
	}

	expense, err := c.expenseRepo.ChangeGoal(*toUpdate, params.GoalID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(map[string]any{"error": "goal not found"})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(map[string]any{"error": "internal server error"})
		return
	}

	encoder.Encode(expense)
}

func (c *ExpenseController) Delete(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(map[string]any{"error": "invalid expense id"})
		return
	}

	err = c.expenseRepo.Delete(uint(id))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		encoder.Encode(map[string]any{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
