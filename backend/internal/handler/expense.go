package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/gofiber/fiber/v2"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/util"
)

type ExpenseHandler struct {
	expenseRepo domain.ExpenseRepo
	goalRepo    domain.GoalRepo
	salaryRepo  domain.SalaryRepo
}

var expenseCreateSchema = z.Struct(z.Schema{
	"name":   z.String().Trim().Min(2, z.Message("name must contain at least 2 characters")).Required(),
	"value":  z.Float().GTE(0.01, z.Message("value must be greater than 0.01")).Required(),
	"date":   z.Time(z.Time.Format(util.ApiDateLayout)).Required(),
	"goalID": z.Int().Required(),
})

var expenseUpdateSchema = z.Struct(z.Schema{
	"name":  z.String().Trim().Min(2, z.Message("name must contain at least 2 characters")).Optional(),
	"value": z.Float().GTE(0.01, z.Message("value must be greater than 0.01")).Optional(),
	"date":  z.Time(z.Time.Format(util.ApiDateLayout)).Optional(),
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
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "query must be present and have at least 2 characters"})
	}

	userID := util.GetUserIDFromCtx(c)
	names := h.expenseRepo.FindMatchingNames(query, userID)

	return c.Status(http.StatusOK).JSON(names)
}

func (h *ExpenseHandler) GetSummary(c *fiber.Ctx) error {
	date := time.Now()

	if queryDate := c.Query("date"); queryDate != "" {
		parsedDate, err := time.Parse(util.ApiDateLayout, queryDate)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid date"})
		}

		date = parsedDate
	}

	userID := util.GetUserIDFromCtx(c)
	summary := h.expenseRepo.GetSummary(date, userID, h.goalRepo, h.salaryRepo)
	return c.Status(http.StatusOK).JSON(summary)
}

func (h *ExpenseHandler) Create(c *fiber.Ctx) error {
	var params struct {
		Name   string
		Value  float64
		Date   time.Time
		GoalID int `zog:"goal_id"`
	}

	if errs := util.ParseZodSchema(expenseCreateSchema, c.Body(), &params); errs != nil {
		return handleZodError(c, errs)
	}

	userID := util.GetUserIDFromCtx(c)

	toCreate := domain.Expense{
		Name:   params.Name,
		Value:  int64(params.Value * 100),
		Date:   params.Date,
		GoalID: uint(params.GoalID),
	}

	expense, err := h.expenseRepo.Create(toCreate, userID, h.goalRepo)
	if err != nil {
		return handleError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(expense.View())
}

func (h *ExpenseHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid expense id"})
	}

	var params struct {
		Name  string    `json:"name"`
		Value float64   `json:"value"`
		Date  time.Time `json:"date"`
	}

	if errs := util.ParseZodSchema(expenseUpdateSchema, c.Body(), &params); errs != nil {
		return handleZodError(c, errs)
	}

	userID := util.GetUserIDFromCtx(c)

	toUpdate, err := h.expenseRepo.Get(uint(id), userID)
	if err != nil {
		return handleError(c, err)
	}

	util.UpdateIfNotZero(&toUpdate.Name, params.Name)
	util.UpdateIfNotZero(&toUpdate.Value, int64(params.Value*100))
	util.UpdateIfNotZero(&toUpdate.Date, params.Date)

	expense, err := h.expenseRepo.Update(toUpdate)
	if err != nil {
		return handleError(c, err)
	}

	return c.Status(http.StatusOK).JSON(expense.View())
}

func (h *ExpenseHandler) UpdateGoal(c *fiber.Ctx) error {
	var params struct {
		GoalID uint `json:"goal_id"`
	}
	if err := json.Unmarshal(c.Body(), &params); err != nil {
		return InvalidJSONBody(c, err)
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid expense id"})
	}

	userID := util.GetUserIDFromCtx(c)

	toUpdate, err := h.expenseRepo.Get(uint(id), userID)
	if err != nil {
		return handleError(c, err)
	}

	expense, err := h.expenseRepo.ChangeGoal(toUpdate, params.GoalID, userID, h.goalRepo)
	if err != nil {
		return handleError(c, err)
	}

	return c.Status(http.StatusOK).JSON(expense.View())
}

func (h *ExpenseHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid expense id"})
	}

	userID := util.GetUserIDFromCtx(c)

	err = h.expenseRepo.Delete(uint(id), userID)
	if err != nil {
		return handleError(c, err)
	}

	return c.Status(http.StatusNoContent).Send(nil)
}
