package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/util"
)

type SalaryHandler struct {
	salaryRepo domain.SalaryRepo
}

func NewSalaryHandler(salaryRepo domain.SalaryRepo) SalaryHandler {
	return SalaryHandler{salaryRepo: salaryRepo}
}

func (h *SalaryHandler) Get(c *fiber.Ctx) error {
	salary := h.salaryRepo.Get()
	return c.Status(http.StatusOK).JSON(util.M{"amount": salary.Amount})
}
