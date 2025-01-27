package config

import (
	"sync"

	"github.com/joaopsramos/fincon/internal/domain"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var ConnectAndSetup = func(path string) *gorm.DB {
	return sync.OnceValue(func() *gorm.DB {
		return connectAndSetup(path)
	})()
}

func connectAndSetup(path string) *gorm.DB {
	db := connect(path)
	setup(db)

	return db
}

func connect(path string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	return db
}

func setup(db *gorm.DB) {
	db.AutoMigrate(&domain.Goal{})
	db.AutoMigrate(&domain.Salary{})
	db.AutoMigrate(&domain.Expense{})
}
