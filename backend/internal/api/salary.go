package api

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/joaopsramos/fincon/internal/util"
)

func (a *Api) GetSalary(c *fiber.Ctx) error {
	userID := util.GetUserIDFromCtx(c)
	salary := a.salaryRepo.Get(userID)
	return c.Status(http.StatusOK).JSON(salary.View())
}
