package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

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
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/api", func(r chi.Router) {
		r.Get("/salary", salaryController.Get)

		r.Route("/expenses", func(r chi.Router) {
			r.Post("/", expenseController.Create)
			r.Patch("/{id}", expenseController.Update)
			r.Delete("/{id}", expenseController.Delete)
			r.Patch("/{id}/update-goal", expenseController.UpdateGoal)
			r.Get("/summary", expenseController.GetSummary)
		})

		r.Route("/goals", func(r chi.Router) {
			r.Get("/", goalController.Index)
			r.Get("/{id}/expenses", goalController.GetExpenses)
		})
	})

	slog.Info("Listening on port 3000")

	if err := http.ListenAndServe(":3000", r); err != nil {
		panic(err)
	}
}
