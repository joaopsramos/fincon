package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/util"
)

type GoalController struct {
	repo        domain.GoalRepo
	expenseRepo domain.ExpenseRepo
}

func NewGoalController(repo domain.GoalRepo, expenseRepo domain.ExpenseRepo) GoalController {
	return GoalController{repo: repo, expenseRepo: expenseRepo}
}

func (c *GoalController) Index(ctx fiber.Ctx) error {
	goals := c.repo.All()
	return ctx.Status(http.StatusOK).JSON(goals)
}

func (c *GoalController) GetExpenses(ctx fiber.Ctx) error {
	query := ctx.Queries()
	now := time.Now()
	year, month, _ := now.Date()

	if queryYear, ok := query["year"]; ok {
		parsedYear, err := strconv.Atoi(queryYear)
		if err != nil || parsedYear < 1 {
			return ctx.Status(http.StatusBadRequest).JSON(util.M{"error": "invalid year"})
		}

		year = parsedYear
	}

	if queryMonth, ok := query["month"]; ok {
		parsedMonth, err := strconv.Atoi(queryMonth)
		if err != nil || parsedMonth < 1 || parsedMonth > 12 {
			return ctx.Status(http.StatusBadRequest).JSON(util.M{"error": "invalid month"})
		}

		month = time.Month(parsedMonth)
	}

	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil || id < 1 {
		return ctx.Status(http.StatusBadRequest).JSON(util.M{"error": "invalid goal id"})
	}

	expenses := c.expenseRepo.AllByGoalID(uint(id), year, month)
	return ctx.Status(http.StatusOK).JSON(expenses)
}
