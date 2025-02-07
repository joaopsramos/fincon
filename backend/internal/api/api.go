package api

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joaopsramos/fincon/internal/handler"
	"github.com/joaopsramos/fincon/internal/repository"
	"github.com/joaopsramos/fincon/internal/util"
	"gorm.io/gorm"
)

type Api struct {
	Router *fiber.App

	logger         *slog.Logger
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

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return &Api{
		Router:         newFiber(),
		logger:         logger,
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
	api := a.Router.Group("/api")

	api.Post("/users", limiter.New(limiter.Config{
		Max:                5,
		Expiration:         1 * time.Hour,
		SkipFailedRequests: true,
		LimitReached:       a.limitReached,
	}), a.userHandler.Create)

	api.Post("/sessions", limiter.New(limiter.Config{
		Max:               10,
		Expiration:        5 * time.Minute,
		LimiterMiddleware: limiter.SlidingWindow{},
		LimitReached:      a.limitReached,
	}), a.userHandler.Login)

	api.Use(a.userHandler.ValidateTokenMiddleware())
	api.Use(a.userHandler.PutUserIDMiddleware())

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

func (a *Api) limitReached(c *fiber.Ctx) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(util.M{"error": "too many requests"})
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
