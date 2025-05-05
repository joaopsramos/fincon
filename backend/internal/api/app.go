package api

import (
	"context"
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
	"github.com/google/uuid"
	"github.com/honeybadger-io/honeybadger-go"
	"github.com/joaopsramos/fincon/internal/auth"
	"github.com/joaopsramos/fincon/internal/mail"
	"github.com/joaopsramos/fincon/internal/repository"
	"github.com/joaopsramos/fincon/internal/service"
	"github.com/joaopsramos/fincon/internal/util"
	"gorm.io/gorm"
)

type App struct {
	Router *chi.Mux
	logger *slog.Logger

	userHandler    *UserHandler
	salaryHandler  *SalaryHandler
	goalHandler    *GoalHandler
	expenseHandler *ExpenseHandler
}

func NewApp(db *gorm.DB) *App {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	userRepo := repository.NewPostgresUser(db)
	salaryRepo := repository.NewPostgresSalary(db)
	goalRepo := repository.NewPostgresGoal(db)
	expenseRepo := repository.NewPostgresExpense(db)

	baseHandler := NewBaseHandler(logger)
	mailer := mail.NewMailer()

	userService := service.NewUserService(userRepo, mailer)
	salaryService := service.NewSalaryService(salaryRepo)
	goalService := service.NewGoalService(goalRepo)
	expenseService := service.NewExpenseService(expenseRepo, goalRepo, salaryRepo)

	return &App{
		Router: chi.NewRouter(),
		logger: logger,

		userHandler:    NewUserHandler(baseHandler, userService),
		salaryHandler:  NewSalaryHandler(baseHandler, salaryService),
		goalHandler:    NewGoalHandler(baseHandler, goalService, expenseService),
		expenseHandler: NewExpenseHandler(baseHandler, expenseService),
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
		w.WriteHeader(http.StatusNotFound)
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
	tokenAuth := auth.NewTokenAuth()

	a.Router.Route("/api", func(r chi.Router) {
		// Public routes
		a.userHandler.RegisterRoutes(r)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator(tokenAuth))
			r.Use(a.PutUserIDMiddleware)

			a.salaryHandler.RegisterRoutes(r)
			a.goalHandler.RegisterRoutes(r)
			a.expenseHandler.RegisterRoutes(r)
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

func (a *App) PutUserIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			panic(err)
		}

		sub, ok := claims["sub"].(string)
		if !ok {
			panic("failed to get subject from token")
		}

		userID, err := uuid.Parse(sub)
		if err != nil {
			panic(err)
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
