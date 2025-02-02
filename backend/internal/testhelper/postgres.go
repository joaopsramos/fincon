package testhelper

import (
	"os"
	"path"
	"testing"

	"github.com/joaopsramos/fincon/internal/config"
	"gorm.io/gorm"
)

func NewTestPostgresTx(t *testing.T) *gorm.DB {
	os.Setenv("APP_ENV", "test")
	config.LoadEnv(path.Join("..", ".."))

	db := config.NewPostgresConn(config.PostgresDSNFromEnv())

	tx := db.Begin()

	t.Cleanup(func() {
		tx.Rollback()
	})

	return tx
}
