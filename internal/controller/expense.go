package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ExpenseController struct {
	expenseRepo domain.ExpenseRepo
	goalRepo    domain.GoalRepo
	salaryRepo  domain.SalaryRepo
}

func NewExpenseController(
	expenseRepo domain.ExpenseRepo,
	goalRepo domain.GoalRepo,
	salaryRepo domain.SalaryRepo,
) ExpenseController {
	return ExpenseController{expenseRepo: expenseRepo, goalRepo: goalRepo, salaryRepo: salaryRepo}
}

func (c *ExpenseController) GetSummary(ctx echo.Context) error {
	date := time.Now()

	if queryDate := ctx.QueryParam("date"); queryDate != "" {
		parsedDate, err := time.Parse("2006-01-02", queryDate)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]any{"error": "invalid date"})
		}

		date = parsedDate
	}

	summary := c.expenseRepo.GetSummary(date, c.goalRepo, c.salaryRepo)
	return ctx.JSON(http.StatusOK, summary)
}

func (c *ExpenseController) Create(ctx echo.Context) error {
	var params struct {
		Name   string  `json:"name"`
		Value  float64 `json:"value"`
		Date   string  `json:"date"`
		GoalID uint    `json:"goal_id"`
	}
	json.NewDecoder(ctx.Request().Body).Decode(&params)

	date, err := time.Parse("02/01/2006", params.Date)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"error": "invalid date"})
	}

	toCreate := domain.Expense{
		Name:   params.Name,
		Value:  int64(params.Value * 100),
		Date:   date,
		GoalID: params.GoalID,
	}

	expense, err := c.expenseRepo.Create(toCreate, c.goalRepo)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"error": "goal not found"})
	} else if err != nil {
		panic(err)
	}

	return ctx.JSON(http.StatusCreated, expense)
}

func (c *ExpenseController) Update(ctx echo.Context) error {
	var params struct {
		Name  string  `json:"name"`
		Value float64 `json:"value"`
		Date  string  `json:"date"`
	}
	json.NewDecoder(ctx.Request().Body).Decode(&params)

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"error": "invalid expense id"})
	}

	date, err := time.Parse("02/01/2006", params.Date)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"error": "invalid date"})
	}

	toUpdate, err := c.expenseRepo.Get(uint(id))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"error": "expense not found"})
	} else if err != nil {
		panic(err)
	}

	toUpdate.Name = params.Name
	toUpdate.Value = int64(params.Value * 100)
	toUpdate.Date = date

	expense, err := c.expenseRepo.Update(*toUpdate)
	if err != nil {
		panic(err)
	}

	return ctx.JSON(http.StatusOK, expense)
}

func (c *ExpenseController) UpdateGoal(ctx echo.Context) error {
	var params struct {
		GoalID uint `json:"goal_id"`
	}
	json.NewDecoder(ctx.Request().Body).Decode(&params)

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"error": "invalid expense id"})
	}

	toUpdate, err := c.expenseRepo.Get(uint(id))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"error": "expense not found"})
	} else if err != nil {
		panic(err)
	}

	expense, err := c.expenseRepo.ChangeGoal(*toUpdate, params.GoalID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"error": "goal not found"})
	} else if err != nil {
		panic(err)
	}

	return ctx.JSON(http.StatusOK, expense)
}

func (c *ExpenseController) Delete(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"error": "invalid expense id"})
	}

	err = c.expenseRepo.Delete(uint(id))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	return ctx.NoContent(http.StatusNoContent)
}
