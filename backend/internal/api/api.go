package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/joaopsramos/fincon/internal/controller"
	"github.com/joaopsramos/fincon/internal/repository"
	"gorm.io/gorm"
)

type Api struct {
	Router *fiber.App

	salaryController  controller.SalaryController
	goalController    controller.GoalController
	expenseController controller.ExpenseController
}

func NewApi(db *gorm.DB) *Api {
	salaryRepo := repository.NewPostgresSalary(db)
	goalRepo := repository.NewPostgresGoal(db)
	expenseRepo := repository.NewPostgresExpense(db)

	salaryController := controller.NewSalaryController(salaryRepo)
	goalController := controller.NewGoalController(goalRepo, expenseRepo)
	expenseController := controller.NewExpenseController(expenseRepo, goalRepo, salaryRepo)

	return &Api{
		Router:            fiber.New(),
		salaryController:  salaryController,
		goalController:    goalController,
		expenseController: expenseController,
	}
}

func (a *Api) Setup() {
	a.SetupMiddlewares()
	a.SetupRoutes()
}

func (a *Api) Listen() error {
	slog.Info("Listening on port 4000")

	return a.Router.Listen(":4000")
}

func (a *Api) SetupMiddlewares() {
	a.Router.Use(logger.New())
	a.Router.Use(recover.New())
	a.Router.Use(cors.New())
}

func (a *Api) SetupRoutes() {
	api := a.Router.Group("/api")

	api.Get("/salary", a.salaryController.Get)

	api.Post("/expenses", a.expenseController.Create)
	api.Patch("/expenses/:id", a.expenseController.Update)
	api.Delete("/expenses/:id", a.expenseController.Delete)
	api.Patch("/expenses/:id/update-goal", a.expenseController.UpdateGoal)
	api.Get("/expenses/summary", a.expenseController.GetSummary)

	api.Get("/goals", a.goalController.Index)
	api.Get("/goals/:id/expenses", a.goalController.GetExpenses)
}
