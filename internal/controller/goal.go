package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/labstack/echo/v4"
)

type GoalController struct {
	repo        domain.GoalRepo
	expenseRepo domain.ExpenseRepo
}

func NewGoalController(repo domain.GoalRepo, expenseRepo domain.ExpenseRepo) GoalController {
	return GoalController{repo: repo, expenseRepo: expenseRepo}
}

func (c *GoalController) Index(ctx echo.Context) error {
	goals := c.repo.All()
	return ctx.JSON(http.StatusOK, goals)
}

func (c *GoalController) GetExpenses(ctx echo.Context) error {
	query := ctx.QueryParams()
	now := time.Now()
	year, month, _ := now.Date()

	queryYear := query.Get("year")
	queryMonth := query.Get("month")

	if queryYear != "" {
		parsedYear, err := strconv.Atoi(queryYear)
		if err != nil || parsedYear < 1 {
			return ctx.JSON(http.StatusBadRequest, map[string]any{"error": "invalid year"})
		}

		year = parsedYear
	}

	if queryMonth != "" {
		parsedMonth, err := strconv.Atoi(queryMonth)
		if err != nil || parsedMonth < 1 || parsedMonth > 12 {
			return ctx.JSON(http.StatusBadRequest, map[string]any{"error": "invalid month"})
		}

		month = time.Month(parsedMonth)
	}

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id < 1 {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"error": "invalid goal id"})
	}

	expenses := c.expenseRepo.AllByGoalID(uint(id), year, month)
	return ctx.JSON(http.StatusOK, expenses)
}
