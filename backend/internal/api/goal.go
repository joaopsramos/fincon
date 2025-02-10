package api

import (
	"encoding/json"
	"fmt"
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

func (a *Api) UpdateGoals(c *fiber.Ctx) error {
	type paramsType struct {
		ID         int `json:"id"`
		Percentage int `json:"percentage"`
	}
	var params []paramsType

	err := json.Unmarshal(c.Body(), &params)
	if err != nil {
		return a.InvalidJSONBody(c, err)
	}

	if len(params) < len(domain.DefaulGoalPercentages()) {
		return c.
			Status(http.StatusBadRequest).
			JSON(util.M{"error": "one or more goals are missing"})
	}

	percentageSum := 0
	for _, p := range params {
		if p.Percentage < 0 || p.Percentage > 100 {
			return c.
				Status(http.StatusBadRequest).
				JSON(util.M{"error": fmt.Sprintf("invalid percentage for goal id %d, it must be greater than or equal to 0 and less then or equal to 100", p.ID)})
		}

		percentageSum += p.Percentage
	}

	if percentageSum != 100 {
		return c.
			Status(http.StatusBadRequest).
			JSON(util.M{"error": "the sum of all percentages must be equal to 100"})
	}

	userID := util.GetUserIDFromCtx(c)
	goals := a.goalRepo.All(userID)

	paramsById := make(map[int]paramsType, len(params))
	for _, p := range params {
		paramsById[p.ID] = p
	}

	for i, g := range goals {
		p, exists := paramsById[int(g.ID)]
		if !exists {
			return c.
				Status(http.StatusBadRequest).
				JSON(util.M{"error": fmt.Sprintf("missing goal with id %d", g.ID)})
		}

		goals[i].Percentage = uint(p.Percentage)
	}

	if err := a.goalRepo.UpdateAll(goals); err != nil {
		return a.HandleError(c, err)
	}

	goalViews := make([]domain.GoalView, 0, len(goals))
	for _, g := range goals {
		goalViews = append(goalViews, g.View())
	}

	return c.Status(http.StatusOK).JSON(goalViews)
}
