package controller

import (
	"encoding/json"
	"net/http"

	"github.com/joaopsramos/fincon/internal/domain"
)

type SalaryController struct {
	salaryRepo domain.SalaryRepo
}

func NewSalaryController(salaryRepo domain.SalaryRepo) SalaryController {
	return SalaryController{salaryRepo: salaryRepo}
}

func (c *SalaryController) Get(w http.ResponseWriter, r *http.Request) {
	salary := c.salaryRepo.Get()
	json.NewEncoder(w).Encode(map[string]any{"amount": salary.Amount})
}
