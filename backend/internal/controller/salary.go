package controller

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/joaopsramos/fincon/internal/domain"
)

type SalaryController struct {
	salaryRepo domain.SalaryRepo
}

func NewSalaryController(salaryRepo domain.SalaryRepo) SalaryController {
	return SalaryController{salaryRepo: salaryRepo}
}

func (c *SalaryController) Get(ctx fiber.Ctx) error {
	salary := c.salaryRepo.Get()
	return ctx.Status(http.StatusOK).JSON(map[string]any{"amount": salary.Amount})
}
