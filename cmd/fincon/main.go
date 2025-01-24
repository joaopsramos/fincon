package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	goalController := controller.NewGoalController(goalRepo, expenseRepo)
	expenseController := controller.NewExpenseController(expenseRepo, goalRepo, salaryRepo)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/api/salary", salaryController.Get)
	r.Get("/api/expenses/summary", expenseController.GetSummary)
	r.Get("/api/goals", goalController.Index)
	r.Get("/api/goals/{id}/expenses", goalController.GetExpenses)

	slog.Info("Listening on port 3000")

	if err := http.ListenAndServe(":3000", r); err != nil {
		panic(err)
	}
}
