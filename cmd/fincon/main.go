package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/joaopsramos/fincon/internal/config"
	"github.com/joaopsramos/fincon/internal/controller"
	"github.com/joaopsramos/fincon/internal/repository"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")

	db := config.ConnectAndSetup(os.Getenv("SQLITE_PATH"))
	salaryRepo := repository.NewSQLiteSalary(db)
	goalRepo := repository.NewSQLiteGoal(db)
	expenseRepo := repository.NewSQLiteExpense(db)

	salaryController := controller.NewSalaryController(salaryRepo)
	expenseController := controller.NewExpenseController(expenseRepo, goalRepo, salaryRepo)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/salary", salaryController.Get)
	mux.HandleFunc("GET /api/expenses/summary", expenseController.GetSummary)

	slog.Info("Listening on port 3000")

	if err := http.ListenAndServe(":3000", mux); err != nil {
		panic(err)
	}
}
