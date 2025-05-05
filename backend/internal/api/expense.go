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

type ExpenseHandler struct {
	*BaseHandler
	expenseService service.ExpenseService
}

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

func NewExpenseHandler(baseHandler *BaseHandler, expenseService service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{
		BaseHandler:    baseHandler,
		expenseService: expenseService,
	}
}

func (h *ExpenseHandler) RegisterRoutes(r chi.Router) {
	r.Post("/expenses", h.Create)
	r.Patch("/expenses/{id}", h.Update)
	r.Delete("/expenses/{id}", h.Delete)
	r.Patch("/expenses/{id}/update-goal", h.UpdateGoal)
	r.Get("/expenses/summary", h.GetSummary)
	r.Get("/expenses/matching-names", h.FindSuggestions)
}

func (h *ExpenseHandler) FindSuggestions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if len(query) < 2 {
		h.sendError(w, http.StatusBadRequest, "query must be present and have at least 2 characters")
		return
	}

	userID := h.getUserIDFromCtx(r)
	names, err := h.expenseService.FindMatchingNames(r.Context(), query, userID)
	if err != nil {
		h.HandleError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, names)
}

func (h *ExpenseHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	date := time.Now()

	if queryDate := r.URL.Query().Get("date"); queryDate != "" {
		parsedDate, err := time.Parse(util.ApiDateLayout, queryDate)
		if err != nil {
			h.sendError(w, http.StatusBadRequest, "invalid date")
			return
		}

		date = parsedDate
	}

	userID := h.getUserIDFromCtx(r)
	summary, err := h.expenseService.GetSummary(r.Context(), date, userID)
	if err != nil {
		h.HandleError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, summary)
}

func (h *ExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         string
		Value        float64
		Date         time.Time
		GoalID       int `zog:"goal_id"`
		Installments int
	}

	if errs := util.ParseZodSchema(expenseCreateSchema, r.Body, &params); errs != nil {
		h.HandleZodError(w, errs)
		return
	}

	userID := h.getUserIDFromCtx(r)

	dto := service.CreateExpenseDTO{
		Name:         params.Name,
		Value:        params.Value,
		Date:         params.Date,
		GoalID:       params.GoalID,
		Installments: params.Installments,
	}

	expenses, err := h.expenseService.Create(r.Context(), dto, userID)
	if err != nil {
		h.HandleError(w, err)
		return
	}

	var expenseDTOs []domain.ExpenseDTO
	for _, e := range expenses {
		expenseDTOs = append(expenseDTOs, e.ToDTO())
	}

	h.sendJSON(w, http.StatusCreated, util.M{"data": expenseDTOs})
}

func (h *ExpenseHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	var params struct {
		Name   string    `json:"name"`
		Value  float64   `json:"value"`
		Date   time.Time `json:"date"`
		GoalID int       `zog:"goal_id"`
	}

	if errs := util.ParseZodSchema(expenseUpdateSchema, r.Body, &params); errs != nil {
		h.HandleZodError(w, errs)
		return
	}

	dto := service.UpdateExpenseDTO{
		Name:   params.Name,
		Value:  params.Value,
		Date:   params.Date,
		GoalID: params.GoalID,
	}

	expense, err := h.expenseService.UpdateByID(r.Context(), uint(id), dto, h.getUserIDFromCtx(r))
	if err != nil {
		h.HandleError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, expense.ToDTO())
}

func (h *ExpenseHandler) UpdateGoal(w http.ResponseWriter, r *http.Request) {
	var params struct {
		GoalID uint `json:"goal_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		h.InvalidJSONBody(w, err)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	userID := h.getUserIDFromCtx(r)

	expense, err := h.expenseService.Get(r.Context(), uint(id), userID)
	if err != nil {
		h.HandleError(w, err)
		return
	}

	err = h.expenseService.ChangeGoal(r.Context(), expense, params.GoalID, userID)
	if err != nil {
		h.HandleError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, expense.ToDTO())
}

func (h *ExpenseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	userID := h.getUserIDFromCtx(r)

	err = h.expenseService.Delete(r.Context(), uint(id), userID)
	if err != nil {
		h.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
