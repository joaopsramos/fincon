package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v3/log"
	"github.com/joaopsramos/fincon/internal/config"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joho/godotenv"
)

func main() {
	if os.Getenv("APP_ENV") == "test" {
		godotenv.Load(".env.test")
	} else {
		godotenv.Load(".env")
	}

	dns := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASS"),
	)

	db := config.NewPGConn(dns)

	log.Info("Creating datbase...")
	db.Exec("CREATE DATABASE " + os.Getenv("POSTGRES_DB"))

	db = config.NewPGConn(fmt.Sprintf("%s dbname=%s", dns, os.Getenv("POSTGRES_DB")))

	log.Info("Auto migrating...")
	db.AutoMigrate(&domain.Goal{}, &domain.Salary{}, &domain.Expense{})
}
