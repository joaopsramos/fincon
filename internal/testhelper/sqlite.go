package testhelper

import (
	"os"

	"github.com/joaopsramos/fincon/internal/config"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func NewTestSQLiteDB() *gorm.DB {
	os.Setenv("APP_ENV", "test")
	godotenv.Load(".env.test")
	return config.ConnectAndSetup()
}
