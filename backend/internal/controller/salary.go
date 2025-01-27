package controller

import (
	"net/http"

	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/labstack/echo/v4"
)

type SalaryController struct {
	salaryRepo domain.SalaryRepo
}

func NewSalaryController(salaryRepo domain.SalaryRepo) SalaryController {
	return SalaryController{salaryRepo: salaryRepo}
}

func (c *SalaryController) Get(ctx echo.Context) error {
	salary := c.salaryRepo.Get()
	return ctx.JSON(http.StatusOK, map[string]any{"amount": salary.Amount})
}
