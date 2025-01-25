package domain

type Salary struct {
	Amount int64 `json:"amount"`
}

type SalaryRepo interface {
	Get() Salary
}
