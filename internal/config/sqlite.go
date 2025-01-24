package config

import (
	"os"
	"path"
	"sync"

	"github.com/joaopsramos/fincon/internal/domain"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var ConnectAndSetup = sync.OnceValue(connectAndSetup)

func connectAndSetup() *gorm.DB {
	db := connect()
	setup(db)

	return db
}

func connect() *gorm.DB {
	path := path.Join("..", "..", os.Getenv("SQLITE_PATH"))
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
