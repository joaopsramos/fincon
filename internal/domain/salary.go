package domain

type Salary struct {
	Amount int64 `json:"amount"`
}

type SalaryRepository interface {
	Get() Salary
}
