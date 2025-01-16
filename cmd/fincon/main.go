package main

import (
	"net/http"
	"os"

	"github.com/joaopsramos/fincon/internal/config"
	"github.com/joaopsramos/fincon/internal/controller"
	"github.com/joaopsramos/fincon/internal/repository"
	"github.com/joho/godotenv"
)

func main() {
	switch env := os.Getenv("APP_ENV"); env {
	case "test":
		godotenv.Load(".env.test")
	default:
		godotenv.Load(".env")
	}

	db := config.ConnectAndSetup()
	r := repository.NewSQLiteSalary(db)
	salaryController := controller.NewSalaryController(r)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/salary", salaryController.Get)

	if err := http.ListenAndServe(":3000", mux); err != nil {
		panic(err)
	}
}
