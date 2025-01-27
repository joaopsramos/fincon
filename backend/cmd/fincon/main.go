package main

import (
	"log/slog"
	"os"

	"github.com/joaopsramos/fincon/internal/config"
	"github.com/joaopsramos/fincon/internal/controller"
	"github.com/joaopsramos/fincon/internal/repository"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	apiG := e.Group("/api")

	apiG.GET("/salary", salaryController.Get)

	apiG.POST("/expenses", expenseController.Create)
	apiG.PATCH("/expenses/:id", expenseController.Update)
	apiG.DELETE("/expenses/:id", expenseController.Delete)
	apiG.PATCH("/expenses/:id/update-goal", expenseController.UpdateGoal)
	apiG.GET("/expenses/summary", expenseController.GetSummary)

	apiG.GET("/goals", goalController.Index)
	apiG.GET("/goals/:id/expenses", goalController.GetExpenses)

	slog.Info("Listening on port 4000")

	e.Logger.Fatal(e.Start(":4000"))
}
