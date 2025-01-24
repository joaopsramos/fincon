package testhelper

import (
	"os"
	"path"

	"github.com/joaopsramos/fincon/internal/config"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func NewTestSQLiteDB() *gorm.DB {
	os.Setenv("APP_ENV", "test")
	godotenv.Load(path.Join("..", "..", ".env.test"))
	return config.ConnectAndSetup(path.Join("..", "..", os.Getenv("SQLITE_PATH")))
}
