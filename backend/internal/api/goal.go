package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/service"
	"github.com/joaopsramos/fincon/internal/util"
)

func (a *Api) AllGoals(c *fiber.Ctx) error {
	userID := util.GetUserIDFromCtx(c)
	goals := a.goalService.All(c.Context(), userID)
	goalDTOs := util.Map(goals, func(g domain.Goal) domain.GoalDTO { return g.ToDTO() })
	return c.Status(http.StatusOK).JSON(goalDTOs)
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

	expenses, err := a.expenseService.AllByGoalID(uint(id), year, month, userID)
	if err != nil {
		return a.HandleError(c, err)
	}

	var expenseDTOs []domain.ExpenseDTO
	for _, e := range expenses {
		expenseDTOs = append(expenseDTOs, e.ToDTO())
	}

	return c.Status(http.StatusOK).JSON(expenseDTOs)
}

func (a *Api) UpdateGoals(c *fiber.Ctx) error {
	var params []struct {
		ID         int `json:"id"`
		Percentage int `json:"percentage"`
	}

	err := json.Unmarshal(c.Body(), &params)
	if err != nil {
		return a.InvalidJSONBody(c, err)
	}

	userID := util.GetUserIDFromCtx(c)

	dtos := make([]service.UpdateGoalDTO, len(params))
	for i, p := range params {
		dtos[i] = service.UpdateGoalDTO{
			ID:         p.ID,
			Percentage: p.Percentage,
		}
	}

	goals, err := a.goalService.UpdateAll(c.Context(), dtos, userID)
	if err != nil {
		return a.HandleError(c, err)
	}

	goalDTOs := util.Map(goals, func(g domain.Goal) domain.GoalDTO { return g.ToDTO() })
	return c.Status(http.StatusOK).JSON(goalDTOs)
}
