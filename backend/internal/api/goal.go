package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/service"
	"github.com/joaopsramos/fincon/internal/util"
)

func (a *Api) AllGoals(w http.ResponseWriter, r *http.Request) {
	userID := a.GetUserIDFromCtx(r)
	goals := a.goalService.All(r.Context(), userID)
	goalDTOs := util.Map(goals, func(g domain.Goal) domain.GoalDTO { return g.ToDTO() })
	a.sendJSON(w, http.StatusOK, goalDTOs)
}

func (a *Api) GetGoalExpenses(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	now := time.Now()
	year, month, _ := now.Date()

	if queryYear := query.Get("year"); queryYear != "" {
		parsedYear, err := strconv.Atoi(queryYear)
		if err != nil || parsedYear < 1 {
			a.sendError(w, http.StatusBadRequest, "invalid year")
			return
		}

		year = parsedYear
	}

	if queryMonth := query.Get("month"); queryMonth != "" {
		parsedMonth, err := strconv.Atoi(queryMonth)
		if err != nil || parsedMonth < 1 || parsedMonth > 12 {
			a.sendError(w, http.StatusBadRequest, "invalid month")
			return
		}

		month = time.Month(parsedMonth)
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		a.sendError(w, http.StatusBadRequest, "invalid goal id")
		return
	}

	userID := a.GetUserIDFromCtx(r)

	expenses, err := a.expenseService.AllByGoalID(r.Context(), uint(id), year, month, userID)
	if err != nil {
		a.HandleError(w, err)
		return
	}

	var expenseDTOs []domain.ExpenseDTO
	for _, e := range expenses {
		expenseDTOs = append(expenseDTOs, e.ToDTO())
	}

	a.sendJSON(w, http.StatusOK, expenseDTOs)
}

func (a *Api) UpdateGoals(w http.ResponseWriter, r *http.Request) {
	var params []struct {
		ID         int `json:"id"`
		Percentage int `json:"percentage"`
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		a.InvalidJSONBody(w, err)
		return
	}

	userID := a.GetUserIDFromCtx(r)

	dtos := make([]service.UpdateGoalDTO, len(params))
	for i, p := range params {
		dtos[i] = service.UpdateGoalDTO{
			ID:         p.ID,
			Percentage: p.Percentage,
		}
	}

	goals, err := a.goalService.UpdateAll(r.Context(), dtos, userID)
	if err != nil {
		a.HandleError(w, err)
		return
	}

	goalDTOs := util.Map(goals, func(g domain.Goal) domain.GoalDTO { return g.ToDTO() })
	a.sendJSON(w, http.StatusOK, goalDTOs)
}
