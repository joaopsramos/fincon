package service

import (
	"github.com/joaopsramos/fincon/internal/domain"
)

type SalaryService struct {
	salaryRepo domain.SalaryRepo
}

type CreateSalaryDTO struct {
	Amount float64
}

func NewSalaryService(salaryRepo domain.SalaryRepo) SalaryService {
	return SalaryService{salaryRepo: salaryRepo}
}

func NewSalary(dto CreateSalaryDTO) domain.Salary {
	return domain.Salary{
		Amount: int64(dto.Amount * 100),
	}
}

func (s *SalaryService) UpdateAmount(salary *domain.Salary, amount float64) error {
	salary.Amount = int64(amount * 100)

	return s.salaryRepo.Update(salary)
}
