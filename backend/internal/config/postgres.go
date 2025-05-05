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

func (c *Config) PostgresDSN() string {
	db := c.Database

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		db.Host,
		db.Port,
		db.User,
		db.Pass,
		db.Name,
	)
}
