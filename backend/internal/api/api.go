package api

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joaopsramos/fincon/internal/handler"
	"github.com/joaopsramos/fincon/internal/repository"
	"github.com/joaopsramos/fincon/internal/util"
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
		Router:         newFiber(),
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
	a.Router.Use(cors.New())
	a.Router.Use(recover.New())
}

func (a *Api) SetupRoutes() {
	a.Router.Post("/api/users", a.userHandler.Create)
	a.Router.Post("/api/sessions", a.userHandler.Login)

	api := a.Router.Group("/api")
	api.Use(a.userHandler.ValidateTokenMiddleware())
	api.Use(a.userHandler.PutUserMiddleware())

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

func newFiber() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError

			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}

			slog.Error(err.Error())

			return ctx.Status(code).JSON(util.M{"error": "internal server error"})
		},
	})
}
