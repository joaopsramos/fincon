package api

import (
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/gofiber/fiber/v2"
	"github.com/joaopsramos/fincon/internal/util"
)

var salaryUpdateSchema = z.Struct(z.Schema{
	"amount": z.Float().GT(0, z.Message("must be greater than 0")).Required(),
})

func (a *Api) GetSalary(c *fiber.Ctx) error {
	userID := util.GetUserIDFromCtx(c)
	salary := util.Must(a.salaryService.Get(c.Context(), userID))

	return c.Status(http.StatusOK).JSON(salary.ToDTO())
}

func (a *Api) UpdateSalary(c *fiber.Ctx) error {
	var params struct {
		Amount float64 `json:"amount"`
	}
	if errs := util.ParseZodSchema(salaryUpdateSchema, c.Body(), &params); errs != nil {
		return a.HandleZodError(c, errs)
	}

	userID := util.GetUserIDFromCtx(c)
	salary := util.Must(a.salaryService.Get(c.Context(), userID))

	if err := a.salaryService.UpdateAmount(c.Context(), salary, params.Amount); err != nil {
		return a.HandleError(c, err)
	}

	return c.Status(http.StatusOK).JSON(salary.ToDTO())
}
