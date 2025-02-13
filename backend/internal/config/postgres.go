package config

import (
	"fmt"
	"log"
	"os"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func PostgresDSNFromEnv() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASS"),
		os.Getenv("POSTGRES_DB"),
	)
}

var NewPostgresConn = func(dsn string) *gorm.DB {
	return sync.OnceValue(func() *gorm.DB {
		logger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				IgnoreRecordNotFoundError: true,
			},
		)

		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger})
		if err != nil {
			panic("failed to connect database")
		}

		return db
	})()
}
