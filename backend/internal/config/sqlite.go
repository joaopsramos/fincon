package config

import (
	"sync"

	"github.com/joaopsramos/fincon/internal/domain"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var NewSQLiteConn = func(path string) *gorm.DB {
	return sync.OnceValue(func() *gorm.DB {
		db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}

		db.AutoMigrate(&domain.Goal{})
		db.AutoMigrate(&domain.Salary{})
		db.AutoMigrate(&domain.Expense{})

		return db
	})()
}
