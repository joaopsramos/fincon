package api

import (
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/joaopsramos/fincon/internal/util"
)

var salaryUpdateSchema = z.Struct(z.Schema{
	"amount": z.Float().GT(0, z.Message("must be greater than 0")).Required(),
})

func (a *Api) GetSalary(w http.ResponseWriter, r *http.Request) {
	userID := a.GetUserIDFromCtx(r)
	salary := util.Must(a.salaryService.Get(r.Context(), userID))
	a.sendJSON(w, http.StatusOK, salary.ToDTO())
}

func (a *Api) UpdateSalary(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Amount float64 `json:"amount"`
	}

	if errs := util.ParseZodSchema(salaryUpdateSchema, r.Body, &params); errs != nil {
		a.HandleZodError(w, errs)
		return
	}

	userID := a.GetUserIDFromCtx(r)
	salary := util.Must(a.salaryService.Get(r.Context(), userID))

	if err := a.salaryService.UpdateAmount(r.Context(), salary, params.Amount); err != nil {
		a.HandleError(w, err)
		return
	}

	a.sendJSON(w, http.StatusOK, salary.ToDTO())
}
