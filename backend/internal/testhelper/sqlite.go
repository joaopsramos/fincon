package testhelper

import (
	"os"
	"path"
	"testing"

	"github.com/joaopsramos/fincon/internal/config"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func NewTestSQLiteTx(t *testing.T) *gorm.DB {
	os.Setenv("APP_ENV", "test")
	godotenv.Load(path.Join("..", "..", ".env.test"))
	db := config.NewSQLiteConn(path.Join("..", "..", os.Getenv("SQLITE_PATH")))

	tx := db.Begin()

	t.Cleanup(func() {
		tx.Rollback()
	})

	return tx
}
