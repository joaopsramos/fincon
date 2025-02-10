package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/util"
)

func (a *Api) AllGoals(c *fiber.Ctx) error {
	userID := util.GetUserIDFromCtx(c)
	goals := a.goalRepo.All(userID)
	return c.Status(http.StatusOK).JSON(goals)
}

func (a *Api) GetGoalExpenses(c *fiber.Ctx) error {
	query := c.Queries()
	now := time.Now()
	year, month, _ := now.Date()

	if queryYear, ok := query["year"]; ok {
		parsedYear, err := strconv.Atoi(queryYear)
		if err != nil || parsedYear < 1 {
			return c.Status(http.StatusBadRequest).JSON(util.M{"error": "invalid year"})
		}

		year = parsedYear
	}

	if queryMonth, ok := query["month"]; ok {
		parsedMonth, err := strconv.Atoi(queryMonth)
		if err != nil || parsedMonth < 1 || parsedMonth > 12 {
			return c.Status(http.StatusBadRequest).JSON(util.M{"error": "invalid month"})
		}

		month = time.Month(parsedMonth)
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return c.Status(http.StatusBadRequest).JSON(util.M{"error": "invalid goal id"})
	}

	userID := util.GetUserIDFromCtx(c)

	expenses := a.expenseRepo.AllByGoalID(uint(id), year, month, userID)

	var expenseViews []domain.ExpenseView
	for _, e := range expenses {
		expenseViews = append(expenseViews, e.View())
	}

	return c.Status(http.StatusOK).JSON(expenseViews)
}
