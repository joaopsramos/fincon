package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	errs "github.com/joaopsramos/fincon/internal/error"
	"github.com/joaopsramos/fincon/internal/util"
)

func (a *Api) HandleError(c *fiber.Ctx, err error) error {
	if errors.Is(err, errs.ErrNotFound{}) {
		return c.Status(http.StatusNotFound).JSON(util.M{"error": err.Error()})
	}

	if errors.Is(err, errs.ErrValidation{}) {
		return c.Status(http.StatusBadRequest).JSON(util.M{"error": err.Error()})
	}

	a.logger.Error(err.Error())
	panic(err)
}

func (a *Api) HandleZodError(c *fiber.Ctx, err map[string]any) error {
	return c.Status(http.StatusBadRequest).JSON(err)
}

func (a *Api) InvalidJSONBody(c *fiber.Ctx, err error) error {
	if errors.Is(err, &json.InvalidUnmarshalError{}) {
		panic(err)
	}

	return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid json body"})
}
