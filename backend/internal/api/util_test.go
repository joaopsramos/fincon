package api_test

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	z "github.com/Oudwins/zog"
	"github.com/gofiber/fiber/v2"
	"github.com/joaopsramos/fincon/internal/api"
	errs "github.com/joaopsramos/fincon/internal/error"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/joaopsramos/fincon/internal/util"
	"github.com/stretchr/testify/assert"
)

func Test_HandleError(t *testing.T) {
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	api := api.NewApi(tx)
	api.SetupMiddlewares()

	api.Router.Get("/not-found", func(c *fiber.Ctx) error {
		return api.HandleError(c, errs.NewNotFound("some resource"))
	})

	api.Router.Get("/some-error", func(c *fiber.Ctx) error {
		return api.HandleError(c, errors.New("some error"))
	})

	data := []struct {
		name           string
		url            string
		expectedStatus int
		expectedBody   fiber.Map
	}{
		{"not found error", "/not-found", 404, fiber.Map{"error": "some resource not found"}},
		{"any other error", "/some-error", 500, fiber.Map{"error": "internal server error"}},
	}

	for _, d := range data {
		var respBody fiber.Map

		req := httptest.NewRequest("GET", d.url, nil)
		resp, _ := api.Router.Test(req)
		_ = json.NewDecoder(resp.Body).Decode(&respBody)

		assert.Equal(d.expectedStatus, resp.StatusCode)
		assert.Equal(d.expectedBody, respBody)
	}
}

func Test_HandleZodError(t *testing.T) {
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	api := api.NewApi(tx)

	api.SetupMiddlewares()

	api.Router.Post("/some-route", func(c *fiber.Ctx) error {
		schema := z.Struct(z.Schema{
			"name": z.String().Required(),
		})
		errs := util.ParseZodSchema(schema, c.Body(), &struct{ Name string }{})
		return api.HandleZodError(c, errs)
	})

	req := httptest.NewRequest("POST", "/some-route", strings.NewReader(`{"not-name": 1}`))
	resp, _ := api.Router.Test(req)

	var respBody fiber.Map
	_ = json.NewDecoder(resp.Body).Decode(&respBody)

	assert.Equal(400, resp.StatusCode)
	assert.Equal(fiber.Map{"errors": util.M{"name": []any{"is required"}}}, respBody)
}

func Test_InvalidJSONBody(t *testing.T) {
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	api := api.NewApi(tx)

	api.SetupMiddlewares()

	api.Router.Post("/some-route", func(c *fiber.Ctx) error {
		err := json.Unmarshal(c.Body(), &fiber.Map{})
		return api.InvalidJSONBody(c, err)
	})

	req := httptest.NewRequest("POST", "/some-route", strings.NewReader(`{"invalid": json}`))
	resp, _ := api.Router.Test(req)

	var respBody fiber.Map
	_ = json.NewDecoder(resp.Body).Decode(&respBody)

	assert.Equal(400, resp.StatusCode)
	assert.Equal(fiber.Map{"error": "invalid json body"}, respBody)
}
