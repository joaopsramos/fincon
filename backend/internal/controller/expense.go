package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/gofiber/fiber/v3"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/util"
	"gorm.io/gorm"
)

type ExpenseController struct {
	expenseRepo domain.ExpenseRepo
	goalRepo    domain.GoalRepo
	salaryRepo  domain.SalaryRepo
}

var expenseCreateSchema = z.Struct(z.Schema{
	"name":   z.String().Min(2).Required(),
	"value":  z.Float().GTE(1).Required(),
	"date":   z.Time(z.Time.Format("2006-01-02")).Required(),
	"goalID": z.Int().Required(),
})

func NewExpenseController(
	expenseRepo domain.ExpenseRepo,
	goalRepo domain.GoalRepo,
	salaryRepo domain.SalaryRepo,
) ExpenseController {
	return ExpenseController{expenseRepo: expenseRepo, goalRepo: goalRepo, salaryRepo: salaryRepo}
}

func (c *ExpenseController) GetSummary(ctx fiber.Ctx) error {
	date := time.Now()

	if queryDate := ctx.Query("date"); queryDate != "" {
		parsedDate, err := time.Parse("2006-01-02", queryDate)
		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(map[string]any{"error": "invalid date"})
		}

		date = parsedDate
	}

	summary := c.expenseRepo.GetSummary(date, c.goalRepo, c.salaryRepo)
	return ctx.Status(http.StatusOK).JSON(summary)
}

func (c *ExpenseController) Create(ctx fiber.Ctx) error {
	var params struct {
		Name   string
		Value  float64
		Date   time.Time
		GoalID int `zog:"goal_id"`
	}

	if err := util.ParseZodSchema(expenseCreateSchema, ctx.Body(), &params); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(err)
	}

	toCreate := domain.Expense{
		Name:   params.Name,
		Value:  int64(params.Value * 100),
		Date:   params.Date,
		GoalID: uint(params.GoalID),
	}

	expense, err := c.expenseRepo.Create(toCreate, c.goalRepo)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]any{"error": "goal not found"})
	} else if err != nil {
		panic(err)
	}

	return ctx.Status(http.StatusCreated).JSON(expense)
}

func (c *ExpenseController) Update(ctx fiber.Ctx) error {
	var params struct {
		Name  string  `json:"name"`
		Value float64 `json:"value"`
		Date  string  `json:"date"`
	}
	json.Unmarshal(ctx.Body(), &params)

	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]any{"error": "invalid expense id"})
	}

	var date time.Time

	if params.Date != "" {
		date, err = time.Parse("2006-01-02", params.Date)
		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(map[string]any{"error": "invalid date"})
		}
	}

	toUpdate, err := c.expenseRepo.Get(uint(id))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]any{"error": "expense not found"})
	} else if err != nil {
		panic(err)
	}

	toUpdate.Name = params.Name
	toUpdate.Value = int64(params.Value * 100)

	if params.Date != "" {
		toUpdate.Date = date
	}

	expense, err := c.expenseRepo.Update(*toUpdate)
	if err != nil {
		panic(err)
	}

	return ctx.Status(http.StatusOK).JSON(expense)
}

func (c *ExpenseController) UpdateGoal(ctx fiber.Ctx) error {
	var params struct {
		GoalID uint `json:"goal_id"`
	}
	json.Unmarshal(ctx.Body(), &params)

	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]any{"error": "invalid expense id"})
	}

	toUpdate, err := c.expenseRepo.Get(uint(id))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]any{"error": "expense not found"})
	} else if err != nil {
		panic(err)
	}

	expense, err := c.expenseRepo.ChangeGoal(*toUpdate, params.GoalID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]any{"error": "goal not found"})
	} else if err != nil {
		panic(err)
	}

	return ctx.Status(http.StatusOK).JSON(expense)
}

func (c *ExpenseController) Delete(ctx fiber.Ctx) error {
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]any{"error": "invalid expense id"})
	}

	err = c.expenseRepo.Delete(uint(id))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]any{"error": err.Error()})
	}

	return ctx.Status(http.StatusNoContent).Send(nil)
}
