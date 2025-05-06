package testhelper

import (
	"reflect"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/joaopsramos/fincon/internal/domain"
	"gorm.io/gorm"
)

type Factory struct {
	tx    *gorm.DB
	faker *gofakeit.Faker
}

func NewFactory(tx *gorm.DB) *Factory {
	return &Factory{tx: tx, faker: gofakeit.New(0)}
}

func (f *Factory) InsertUser(u ...*domain.User) domain.User {
	return insert(f, u, domain.User{Email: f.faker.Email()})
}

func (f *Factory) InsertSalary(s ...*domain.Salary) domain.Salary {
	return insert(f, s, domain.Salary{Amount: f.faker.Int64()})
}

func (f *Factory) InsertGoal(g ...*domain.Goal) domain.Goal {
	return insert(f, g, domain.Goal{Percentage: f.faker.UintRange(1, 100)})
}

func (f *Factory) InsertExpense(e ...*domain.Expense) domain.Expense {
	return insert(f, e, domain.Expense{Name: f.faker.ProductName(), Value: f.faker.Int64(), Date: f.faker.Date()})
}

func insert[T any](f *Factory, given []*T, fake T) T {
	if len(given) < 1 {
		f.tx.Create(&fake)
		return fake
	}

	for i := range given {
		mergeStructs(fake, given[i])
	}

	f.tx.Create(given)

	var zero T
	return zero
}

func mergeStructs[T any](src T, dst *T) {
	sv := reflect.ValueOf(src)
	dv := reflect.ValueOf(dst).Elem()

	for i := range dv.NumField() {
		if dv.Field(i).IsZero() {
			dv.Field(i).Set(sv.Field(i))
		}
	}
}
