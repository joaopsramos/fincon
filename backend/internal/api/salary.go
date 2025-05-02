package api

import (
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/go-chi/chi/v5"
	"github.com/joaopsramos/fincon/internal/service"
	"github.com/joaopsramos/fincon/internal/util"
)

type SalaryHandler struct {
	*Handler
	salaryService service.SalaryService
}

var salaryUpdateSchema = z.Struct(z.Schema{
	"amount": z.Float().GT(0, z.Message("must be greater than 0")).Required(),
})

func NewSalaryHandler(salaryService service.SalaryService, handler *Handler) *SalaryHandler {
	return &SalaryHandler{
		Handler:       handler,
		salaryService: salaryService,
	}
}

func (h *SalaryHandler) RegisterRoutes(r chi.Router) {
	r.Get("/salary", h.GetSalary)
	r.Patch("/salary", h.UpdateSalary)
}

func (h *SalaryHandler) GetSalary(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromCtx(r)
	salary := util.Must(h.salaryService.Get(r.Context(), userID))
	h.sendJSON(w, http.StatusOK, salary.ToDTO())
}

func (h *SalaryHandler) UpdateSalary(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Amount float64 `json:"amount"`
	}

	if errs := util.ParseZodSchema(salaryUpdateSchema, r.Body, &params); errs != nil {
		h.HandleZodError(w, errs)
		return
	}

	userID := h.getUserIDFromCtx(r)
	salary := util.Must(h.salaryService.Get(r.Context(), userID))

	if err := h.salaryService.UpdateAmount(r.Context(), salary, params.Amount); err != nil {
		h.HandleError(w, err)
		return
	}

	h.sendJSON(w, http.StatusOK, salary.ToDTO())
}
