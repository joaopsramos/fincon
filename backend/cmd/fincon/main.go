package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/joaopsramos/fincon/internal/config"
	"github.com/joaopsramos/fincon/internal/controller"
	"github.com/joaopsramos/fincon/internal/repository"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"

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

	app := App{
		router:            fiber.New(),
		salaryController:  salaryController,
		goalController:    goalController,
		expenseController: expenseController,
	}

	app.setupMiddlewarea()
	app.setupRoutes()

	slog.Info("Listening on port 4000")

	log.Fatal(app.router.Listen(":4000"))
}

type App struct {
	router *fiber.App

	salaryController  controller.SalaryController
	goalController    controller.GoalController
	expenseController controller.ExpenseController
}

func (a *App) setupMiddlewarea() {
	a.router.Use(logger.New())
	a.router.Use(recover.New())
	a.router.Use(cors.New())
}

func (a *App) setupRoutes() {
	api := a.router.Group("/api")

	api.Get("/salary", a.salaryController.Get)

	api.Post("/expenses", a.expenseController.Create)
	api.Patch("/expenses/:id", a.expenseController.Update)
	api.Delete("/expenses/:id", a.expenseController.Delete)
	api.Patch("/expenses/:id/update-goal", a.expenseController.UpdateGoal)
	api.Get("/expenses/summary", a.expenseController.GetSummary)

	api.Get("/goals", a.goalController.Index)
	api.Get("/goals/:id/expenses", a.goalController.GetExpenses)
}
