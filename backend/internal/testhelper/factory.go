package testhelper

import (
	"reflect"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
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
	return insert(f, u, domain.User{ID: uuid.New(), Email: f.faker.Email()})
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

func (f *Factory) InsertUserToken(t ...*domain.UserToken) domain.UserToken {
	return insert(f, t, domain.UserToken{Token: uuid.New(), ExpiresAt: time.Now().UTC().Add(24 * time.Hour)})
}

func insert[T any](f *Factory, given []*T, fake T) T {
	if len(given) < 1 {
		if err := f.tx.Create(&fake).Error; err != nil {
			panic(err)
		}

		return fake
	}

	for i := range given {
		mergeStructs(fake, given[i])
	}

	if err := f.tx.Create(given).Error; err != nil {
		panic(err)
	}

	if len(given) == 1 {
		return *given[0]
	}

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
