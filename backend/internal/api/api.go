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
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/repository"
	"github.com/joaopsramos/fincon/internal/util"
	"gorm.io/gorm"
)

type Api struct {
	Router *fiber.App

	logger      *slog.Logger
	userRepo    domain.UserRepo
	salaryRepo  domain.SalaryRepo
	goalRepo    domain.GoalRepo
	expenseRepo domain.ExpenseRepo
}

func NewApi(db *gorm.DB) *Api {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	userRepo := repository.NewPostgresUser(db)
	salaryRepo := repository.NewPostgresSalary(db)
	goalRepo := repository.NewPostgresGoal(db)
	expenseRepo := repository.NewPostgresExpense(db)

	return &Api{
		Router: newFiber(),
		logger: logger,

		userRepo:    userRepo,
		salaryRepo:  salaryRepo,
		goalRepo:    goalRepo,
		expenseRepo: expenseRepo,
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
	}), a.CreateUser)

	api.Post("/sessions", limiter.New(limiter.Config{
		Max:               10,
		Expiration:        5 * time.Minute,
		LimiterMiddleware: limiter.SlidingWindow{},
		LimitReached:      a.limitReached,
	}), a.UserLogin)

	api.Use(a.ValidateTokenMiddleware())
	api.Use(a.PutUserIDMiddleware())

	api.Get("/salary", a.GetSalary)

	api.Post("/expenses", a.CreateExpense)
	api.Patch("/expenses/:id", a.UpdateExpense)
	api.Delete("/expenses/:id", a.DeleteExpense)
	api.Patch("/expenses/:id/update-goal", a.UpdateExpenseGoal)
	api.Get("/expenses/summary", a.GetSummary)
	api.Get("/expenses/matching-names", a.FindExpenseSuggestions)

	api.Get("/goals", a.AllGoals)
	api.Get("/goals/:id/expenses", a.GetGoalExpenses)
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
