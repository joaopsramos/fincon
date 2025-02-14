package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
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

	tests := []struct {
		name         string
		handler      func(c *fiber.Ctx) error
		expectedCode int
		expectedBody fiber.Map
	}{
		{
			name: "panic error",
			handler: func(c *fiber.Ctx) error {
				panic("test error")
			},
			expectedCode: 500,
			expectedBody: fiber.Map{"error": "internal server error"},
		},
		{
			name: "fiber error",
			handler: func(c *fiber.Ctx) error {
				return fiber.NewError(http.StatusNotFound, "resource not found")
			},
			expectedCode: 404,
			expectedBody: fiber.Map{"error": "resource not found"},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			routePath := fmt.Sprintf("/test-route-%d", i)
			api.Router.Get(routePath, tt.handler)

			req := httptest.NewRequest("GET", routePath, nil)
			resp, _ := api.Router.Test(req)

			var respBody fiber.Map
			_ = json.NewDecoder(resp.Body).Decode(&respBody)

			assert.Equal(tt.expectedCode, resp.StatusCode)
			assert.Equal(tt.expectedBody, respBody)
		})
	}
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
			retry := util.Must(strconv.Atoi(resp.Header.Get("retry-after")))
			assert.InDelta(60, retry, 3)
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
			retry := util.Must(strconv.Atoi(resp.Header.Get("retry-after")))
			assert.InDelta(3600, retry, 3)
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
			retry := util.Must(strconv.Atoi(resp.Header.Get("retry-after")))
			assert.InDelta(300, retry, 3)
			break
		}

		assert.Equal(http.StatusCreated, resp.StatusCode)
	}
}
