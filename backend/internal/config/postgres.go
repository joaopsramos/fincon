package config

import (
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var NewPostgresConn = func(dns string) *gorm.DB {
	return sync.OnceValue(func() *gorm.DB {
		db, err := gorm.Open(postgres.Open(dns), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}

		return db
	})()
}
