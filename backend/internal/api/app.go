package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/go-chi/jwtauth/v5"
	"github.com/honeybadger-io/honeybadger-go"
	"github.com/joaopsramos/fincon/internal/config"
	"github.com/joaopsramos/fincon/internal/errs"
	"github.com/joaopsramos/fincon/internal/repository"
	"github.com/joaopsramos/fincon/internal/service"
	"github.com/joaopsramos/fincon/internal/util"
	"github.com/lestrrat-go/jwx/jwa"
	"gorm.io/gorm"
)

type App struct {
	Router *chi.Mux

	logger *slog.Logger

	userService    service.UserService
	salaryService  service.SalaryService
	goalService    service.GoalService
	expenseService service.ExpenseService
}

func NewApp(db *gorm.DB) *App {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	userRepo := repository.NewPostgresUser(db)
	salaryRepo := repository.NewPostgresSalary(db)
	goalRepo := repository.NewPostgresGoal(db)
	expenseRepo := repository.NewPostgresExpense(db)

	return &App{
		Router: chi.NewRouter(),
		logger: logger,

		userService:    service.NewUserService(userRepo),
		salaryService:  service.NewSalaryService(salaryRepo),
		goalService:    service.NewGoalService(goalRepo),
		expenseService: service.NewExpenseService(expenseRepo, goalRepo, salaryRepo),
	}
}

func (a *App) SetupAll() {
	a.SetupMiddlewares()
	a.SetupRoutes()
}

func (a *App) Listen() error {
	slog.Info("Listening on port 4000")
	return http.ListenAndServe(":4000", honeybadger.Handler(a.Router))
}

func (a *App) SetupMiddlewares() {
	if os.Getenv("APP_ENV") != "test" {
		a.Router.Use(middleware.Logger)
	}

	a.Router.Use(middleware.Recoverer)
	a.Router.Use(a.corsMiddleware())
	a.Router.Use(a.rateLimiter(100, time.Minute))

	a.Router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		a.HandleError(w, errs.NewNotFound("not found"))
	})
}

func (a *App) corsMiddleware() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	})
}

func (a *App) SetupRoutes() {
	tokenAuth := jwtauth.New(jwa.HS256.String(), config.SecretKey(), nil)

	a.Router.Route("/api", func(r chi.Router) {
		// Public routes
		r.With(a.rateLimiter(5, time.Hour)).Post("/users", a.CreateUser)
		r.With(a.rateLimiter(10, 5*time.Minute)).Post("/sessions", a.UserLogin)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator(tokenAuth))
			r.Use(a.PutUserIDMiddleware)

			r.Get("/salary", a.GetSalary)
			r.Patch("/salary", a.UpdateSalary)

			r.Post("/expenses", a.CreateExpense)
			r.Patch("/expenses/{id}", a.UpdateExpense)
			r.Delete("/expenses/{id}", a.DeleteExpense)
			r.Patch("/expenses/{id}/update-goal", a.UpdateExpenseGoal)
			r.Get("/expenses/summary", a.GetSummary)
			r.Get("/expenses/matching-names", a.FindExpenseSuggestions)

			r.Get("/goals", a.AllGoals)
			r.Get("/goals/{id}/expenses", a.GetGoalExpenses)
			r.Post("/goals", a.UpdateGoals)
		})
	})
}

// Helper to send JSON responses
func (a *App) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		a.logger.Error("Failed to encode response", "error", err)
	}
}

// Helper to send error responses
func (a *App) sendError(w http.ResponseWriter, status int, message string) {
	a.sendJSON(w, status, util.M{"error": message})
}

func (a *App) rateLimiter(limit int, windowLength time.Duration) func(http.Handler) http.Handler {
	rateLimiter := httprate.NewRateLimiter(
		limit,
		windowLength,
		httprate.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
			a.sendError(w, http.StatusTooManyRequests, "too many requests")
		}),
	)

	return rateLimiter.Handler
}
