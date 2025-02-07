package api_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/joaopsramos/fincon/internal/api"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/stretchr/testify/assert"
)

func TestApi_ErrorHandler(t *testing.T) {
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	api := api.NewApi(tx)

	api.SetupMiddlewares()

	api.Router.Get("/some-route", func(c *fiber.Ctx) error {
		panic("test error")
	})

	req := httptest.NewRequest("GET", "/some-route", nil)
	resp, _ := api.Router.Test(req)

	var respBody fiber.Map
	json.NewDecoder(resp.Body).Decode(&respBody)

	assert.Equal(500, resp.StatusCode)
	assert.Equal(fiber.Map{"error": "internal server error"}, respBody)
}
