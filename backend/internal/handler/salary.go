package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/util"
)

type SalaryHandler struct {
	salaryRepo domain.SalaryRepo
}

func NewSalaryHandler(salaryRepo domain.SalaryRepo) SalaryHandler {
	return SalaryHandler{salaryRepo: salaryRepo}
}

func (c *SalaryHandler) Get(ctx fiber.Ctx) error {
	salary := c.salaryRepo.Get()
	return ctx.Status(http.StatusOK).JSON(util.M{"amount": salary.Amount})
}
