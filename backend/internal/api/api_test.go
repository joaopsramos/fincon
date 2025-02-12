package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/joaopsramos/fincon/internal/api"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/joaopsramos/fincon/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestApi_ErrorHandler(t *testing.T) {
	t.Parallel()
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
	_ = json.NewDecoder(resp.Body).Decode(&respBody)

	assert.Equal(500, resp.StatusCode)
	assert.Equal(fiber.Map{"error": "internal server error"}, respBody)
}

func TestApi_GobalRateLimiter(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	api := api.NewApi(tx)
	api.SetupMiddlewares()

	api.Router.Get("/some-route", func(c *fiber.Ctx) error {
		return c.Status(http.StatusNoContent).Send(nil)
	})

	for i := 1; i <= 101; i++ {
		req := httptest.NewRequest("GET", "/some-route", nil)
		resp, _ := api.Router.Test(req)

		if i == 101 {
			assert.Equal(http.StatusTooManyRequests, resp.StatusCode)
			assert.Equal("60", resp.Header.Get("retry-after"))
			break
		}

		assert.Equal(http.StatusNoContent, resp.StatusCode)
	}
}

func TestApi_CreateUserRateLimiter(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	api := testhelper.NewTestApi(tx)

	for i := 1; i <= 6; i++ {
		user := util.M{"email": fmt.Sprintf("user-%d@mail.com", i), "password": 12345678, "salary": 1000}
		resp := api.Test(http.MethodPost, "/api/users", user)

		if i == 6 {
			assert.Equal(http.StatusTooManyRequests, resp.StatusCode)
			assert.Equal("3600", resp.Header.Get("retry-after"))
			break
		}

		assert.Equal(http.StatusCreated, resp.StatusCode)
	}
}

func TestApi_CreateSessionRateLimiter(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	api := testhelper.NewTestApi(tx)

	user := util.M{"email": "user@mail.com", "password": 12345678, "salary": 1000}
	resp := api.Test(http.MethodPost, "/api/users", user)
	assert.Equal(http.StatusCreated, resp.StatusCode)

	for i := 1; i <= 11; i++ {
		resp := api.Test(http.MethodPost, "/api/sessions", user)

		if i == 11 {
			assert.Equal(http.StatusTooManyRequests, resp.StatusCode)
			assert.Equal("300", resp.Header.Get("retry-after"))
			break
		}

		assert.Equal(http.StatusCreated, resp.StatusCode)
	}
}
