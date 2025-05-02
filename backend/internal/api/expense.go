package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/go-chi/chi/v5"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/service"
	"github.com/joaopsramos/fincon/internal/util"
)

var expenseCreateSchema = z.Struct(z.Schema{
	"name":         z.String().Trim().Min(2, z.Message("name must contain at least 2 characters")).Required(),
	"value":        z.Float().GTE(0.01, z.Message("value must be greater than or equal to 0.01")).Required(),
	"date":         z.Time(z.Time.Format(util.ApiDateLayout)).Required(),
	"goalID":       z.Int().Required(),
	"installments": z.Int().GTE(1, z.Message("installments must be greater than or equal to 1")).Optional(),
})

var expenseUpdateSchema = z.Struct(z.Schema{
	"name":   z.String().Trim().Min(2, z.Message("name must contain at least 2 characters")).Optional(),
	"value":  z.Float().GTE(0.01, z.Message("value must be greater than 0.01")).Optional(),
	"date":   z.Time(z.Time.Format(util.ApiDateLayout)).Optional(),
	"goalID": z.Int().Optional(),
})

func (a *App) FindExpenseSuggestions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if len(query) < 2 {
		a.sendError(w, http.StatusBadRequest, "query must be present and have at least 2 characters")
		return
	}

	userID := a.GetUserIDFromCtx(r)
	names, err := a.expenseService.FindMatchingNames(r.Context(), query, userID)
	if err != nil {
		a.HandleError(w, err)
		return
	}

	a.sendJSON(w, http.StatusOK, names)
}

func (a *App) GetSummary(w http.ResponseWriter, r *http.Request) {
	date := time.Now()

	if queryDate := r.URL.Query().Get("date"); queryDate != "" {
		parsedDate, err := time.Parse(util.ApiDateLayout, queryDate)
		if err != nil {
			a.sendError(w, http.StatusBadRequest, "invalid date")
			return
		}

		date = parsedDate
	}

	userID := a.GetUserIDFromCtx(r)
	summary, err := a.expenseService.GetSummary(r.Context(), date, userID)
	if err != nil {
		a.HandleError(w, err)
		return
	}

	a.sendJSON(w, http.StatusOK, summary)
}

func (a *App) CreateExpense(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         string
		Value        float64
		Date         time.Time
		GoalID       int `zog:"goal_id"`
		Installments int
	}

	if errs := util.ParseZodSchema(expenseCreateSchema, r.Body, &params); errs != nil {
		a.HandleZodError(w, errs)
		return
	}

	userID := a.GetUserIDFromCtx(r)

	dto := service.CreateExpenseDTO{
		Name:         params.Name,
		Value:        params.Value,
		Date:         params.Date,
		GoalID:       params.GoalID,
		Installments: params.Installments,
	}

	expenses, err := a.expenseService.Create(r.Context(), dto, userID)
	if err != nil {
		a.HandleError(w, err)
		return
	}

	var expenseDTOs []domain.ExpenseDTO
	for _, e := range expenses {
		expenseDTOs = append(expenseDTOs, e.ToDTO())
	}

	a.sendJSON(w, http.StatusCreated, util.M{"data": expenseDTOs})
}

func (a *App) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		a.sendError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	var params struct {
		Name   string    `json:"name"`
		Value  float64   `json:"value"`
		Date   time.Time `json:"date"`
		GoalID int       `zog:"goal_id"`
	}

	if errs := util.ParseZodSchema(expenseUpdateSchema, r.Body, &params); errs != nil {
		a.HandleZodError(w, errs)
		return
	}

	dto := service.UpdateExpenseDTO{
		Name:   params.Name,
		Value:  params.Value,
		Date:   params.Date,
		GoalID: params.GoalID,
	}

	expense, err := a.expenseService.UpdateByID(r.Context(), uint(id), dto, a.GetUserIDFromCtx(r))
	if err != nil {
		a.HandleError(w, err)
		return
	}

	a.sendJSON(w, http.StatusOK, expense.ToDTO())
}

func (a *App) UpdateExpenseGoal(w http.ResponseWriter, r *http.Request) {
	var params struct {
		GoalID uint `json:"goal_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		a.InvalidJSONBody(w, err)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		a.sendError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	userID := a.GetUserIDFromCtx(r)

	expense, err := a.expenseService.Get(r.Context(), uint(id), userID)
	if err != nil {
		a.HandleError(w, err)
		return
	}

	err = a.expenseService.ChangeGoal(r.Context(), expense, params.GoalID, userID)
	if err != nil {
		a.HandleError(w, err)
		return
	}

	a.sendJSON(w, http.StatusOK, expense.ToDTO())
}

func (a *App) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		a.sendError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	userID := a.GetUserIDFromCtx(r)

	err = a.expenseService.Delete(r.Context(), uint(id), userID)
	if err != nil {
		a.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
