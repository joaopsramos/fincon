package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
	errs "github.com/joaopsramos/fincon/internal/error"
	"github.com/joaopsramos/fincon/internal/util"
)

func handleError(c *fiber.Ctx, err error) error {
	if errors.Is(err, errs.ErrNotFound{}) {
		return c.Status(http.StatusNotFound).JSON(util.M{"error": err.Error()})
	}

	slog.Error(err.Error())
	panic(err)
}
