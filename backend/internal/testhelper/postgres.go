package testhelper

import (
	"testing"

	"github.com/joaopsramos/fincon/internal/config"
	"gorm.io/gorm"
)

func NewTestPostgresTx(t testing.TB) *gorm.DB {
	db := config.NewPostgresConn(config.Get().PostgresDSN())
	tx := db.Begin()

	t.Cleanup(func() {
		tx.Rollback()
	})

	return tx
}
