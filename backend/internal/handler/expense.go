package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/gofiber/fiber/v2"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/util"
	"gorm.io/gorm"
)

type ExpenseHandler struct {
	expenseRepo domain.ExpenseRepo
	goalRepo    domain.GoalRepo
	salaryRepo  domain.SalaryRepo
}

var expenseCreateSchema = z.Struct(z.Schema{
	"name":   z.String().Trim().Min(2).Required(),
	"value":  z.Float().GTE(1).Required(),
	"date":   z.Time(z.Time.Format(util.ApiDateLayout)).Required(),
	"goalID": z.Int().Required(),
})

func NewExpenseHandler(
	expenseRepo domain.ExpenseRepo,
	goalRepo domain.GoalRepo,
	salaryRepo domain.SalaryRepo,
) ExpenseHandler {
	return ExpenseHandler{expenseRepo: expenseRepo, goalRepo: goalRepo, salaryRepo: salaryRepo}
}

func (h *ExpenseHandler) FindMatchingNames(c *fiber.Ctx) error {
	query := c.Query("query")
	if len(query) < 2 {
		return c.Status(http.StatusBadRequest).JSON(util.M{"error": "query must be present and have at least 2 characters"})
	}

	user := util.GetUserFromCtx(c)
	names := h.expenseRepo.FindMatchingNames(query, user.ID)

	return c.Status(http.StatusOK).JSON(names)
}

func (h *ExpenseHandler) GetSummary(c *fiber.Ctx) error {
	date := time.Now()

	if queryDate := c.Query("date"); queryDate != "" {
		parsedDate, err := time.Parse(util.ApiDateLayout, queryDate)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(util.M{"error": "invalid date"})
		}

		date = parsedDate
	}

	user := util.GetUserFromCtx(c)
	summary := h.expenseRepo.GetSummary(date, user.ID, h.goalRepo, h.salaryRepo)
	return c.Status(http.StatusOK).JSON(summary)
}

func (h *ExpenseHandler) Create(c *fiber.Ctx) error {
	var params struct {
		Name   string
		Value  float64
		Date   time.Time
		GoalID int `zog:"goal_id"`
	}

	if err := util.ParseZodSchema(expenseCreateSchema, c.Body(), &params); err != nil {
		return c.Status(http.StatusBadRequest).JSON(err)
	}

	toCreate := domain.Expense{
		Name:   params.Name,
		Value:  int64(params.Value * 100),
		Date:   params.Date,
		GoalID: uint(params.GoalID),
	}

	user := util.GetUserFromCtx(c)
	expense, err := h.expenseRepo.Create(toCreate, user.ID, h.goalRepo)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(http.StatusBadRequest).JSON(util.M{"error": "goal not found"})
	} else if err != nil {
		panic(err)
	}

	return c.Status(http.StatusCreated).JSON(expense)
}

func (h *ExpenseHandler) Update(c *fiber.Ctx) error {
	var params struct {
		Name  string  `json:"name"`
		Value float64 `json:"value"`
		Date  string  `json:"date"`
	}
	json.Unmarshal(c.Body(), &params)

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.M{"error": "invalid expense id"})
	}

	var date time.Time

	if params.Date != "" {
		date, err = time.Parse(util.ApiDateLayout, params.Date)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(util.M{"error": "invalid date"})
		}
	}

	user := util.GetUserFromCtx(c)

	toUpdate, err := h.expenseRepo.Get(uint(id), user.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(http.StatusBadRequest).JSON(util.M{"error": "expense not found"})
	} else if err != nil {
		panic(err)
	}

	toUpdate.Name = params.Name
	toUpdate.Value = int64(params.Value * 100)

	if params.Date != "" {
		toUpdate.Date = date
	}

	expense, err := h.expenseRepo.Update(*toUpdate)
	if err != nil {
		panic(err)
	}

	return c.Status(http.StatusOK).JSON(expense)
}

func (h *ExpenseHandler) UpdateGoal(c *fiber.Ctx) error {
	var params struct {
		GoalID uint `json:"goal_id"`
	}
	json.Unmarshal(c.Body(), &params)

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.M{"error": "invalid expense id"})
	}

	user := util.GetUserFromCtx(c)

	toUpdate, err := h.expenseRepo.Get(uint(id), user.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(http.StatusBadRequest).JSON(util.M{"error": "expense not found"})
	} else if err != nil {
		panic(err)
	}

	expense, err := h.expenseRepo.ChangeGoal(*toUpdate, params.GoalID, user.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(http.StatusBadRequest).JSON(util.M{"error": "goal not found"})
	} else if err != nil {
		panic(err)
	}

	return c.Status(http.StatusOK).JSON(expense)
}

func (h *ExpenseHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.M{"error": "invalid expense id"})
	}

	user := util.GetUserFromCtx(c)

	err = h.expenseRepo.Delete(uint(id), user.ID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.M{"error": err.Error()})
	}

	return c.Status(http.StatusNoContent).Send(nil)
}
