package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/joaopsramos/fincon/internal/handler"
	"github.com/joaopsramos/fincon/internal/repository"
	"gorm.io/gorm"
)

type Api struct {
	Router *fiber.App

	userHandler    handler.UserHandler
	salaryHandler  handler.SalaryHandler
	goalHandler    handler.GoalHandler
	expenseHandler handler.ExpenseHandler
}

func NewApi(db *gorm.DB) *Api {
	userRepo := repository.NewPostgresUser(db)
	salaryRepo := repository.NewPostgresSalary(db)
	goalRepo := repository.NewPostgresGoal(db)
	expenseRepo := repository.NewPostgresExpense(db)

	userHandler := handler.NewUserHandler(userRepo)
	salaryHandler := handler.NewSalaryHandler(salaryRepo)
	goalHandler := handler.NewGoalHandler(goalRepo, expenseRepo)
	expenseHandler := handler.NewExpenseHandler(expenseRepo, goalRepo, salaryRepo)

	return &Api{
		Router:         fiber.New(),
		userHandler:    userHandler,
		salaryHandler:  salaryHandler,
		goalHandler:    goalHandler,
		expenseHandler: expenseHandler,
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

	api.Post("/users", a.userHandler.Create)
	api.Post("/sessions", a.userHandler.Login)

	api.Get("/salary", a.salaryHandler.Get)

	api.Post("/expenses", a.expenseHandler.Create)
	api.Patch("/expenses/:id", a.expenseHandler.Update)
	api.Delete("/expenses/:id", a.expenseHandler.Delete)
	api.Patch("/expenses/:id/update-goal", a.expenseHandler.UpdateGoal)
	api.Get("/expenses/summary", a.expenseHandler.GetSummary)
	api.Get("/expenses/matching-names", a.expenseHandler.FindMatchingNames)

	api.Get("/goals", a.goalHandler.Index)
	api.Get("/goals/:id/expenses", a.goalHandler.GetExpenses)
}
