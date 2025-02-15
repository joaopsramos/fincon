package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	originalApi "github.com/joaopsramos/fincon/internal/api"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/joaopsramos/fincon/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestUserHandler_Create(t *testing.T) {
	t.Parallel()
	tx := testhelper.NewTestPostgresTx(t)
	api := testhelper.NewTestApi(tx)

	data := []struct {
		name           string
		body           util.M
		expectedStatus int
		expectedBody   util.M
	}{
		{
			"create user successfully",
			util.M{
				"email":    "test@example.com",
				"password": "password123",
				"salary":   5000.00,
			},
			201,
			util.M{
				"salary": util.M{"amount": float64(5000)},
				"user":   util.M{"email": "test@example.com"},
			},
		},
		{
			"email already in use",
			util.M{
				"email":    "test@example.com",
				"password": "password123",
				"salary":   5000.00,
			},
			409,
			util.M{"error": "email already in use"},
		},
		{
			"invalid email",
			util.M{
				"email":    "invalid-email",
				"password": "password123",
				"salary":   5000.00,
			},
			400,
			util.M{"errors": util.M{"email": []any{"must be valid"}}},
		},
		{
			"password too short",
			util.M{
				"email":    "test@example.com",
				"password": "short",
				"salary":   5000.00,
			},
			400,
			util.M{"errors": util.M{"password": []any{"must contain at least 8 characters"}}},
		},
		{
			"password too large",
			util.M{
				"email":    "test@example.com",
				"password": "ThisIsAReallyLongPasswordThatExceedsTheMaximumLengthOf72CharactersAndShouldFailValidation",
				"salary":   5000.00,
			},
			400,
			util.M{"errors": util.M{"password": []any{"must have at most 72 characters"}}},
		},
		{
			"salary must be greater than 0",
			util.M{
				"email":    "test@example.com",
				"password": "password123",
				"salary":   0,
			},
			400,
			util.M{"errors": util.M{"salary": []any{"must be greater than 0"}}},
		},
		{
			"missing required fields",
			util.M{},
			400,
			util.M{"errors": util.M{
				"email":    []any{"is required"},
				"password": []any{"is required"},
				"salary":   []any{"is required"},
			}},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			a := assert.New(t)

			var respBody util.M

			resp := api.Test(http.MethodPost, "/api/users", d.body)
			api.UnmarshalBody(resp.Body, &respBody)
			a.Equal(d.expectedStatus, resp.StatusCode)

			if d.expectedStatus == 201 {
				a.Equal(d.expectedBody["email"], respBody["email"])
				a.Equal(d.expectedBody["salary"], respBody["salary"])
				a.NotEmpty(respBody["user"].(util.M)["id"])
				a.NotEmpty(respBody["token"])

				// assert valid token
				api2 := originalApi.NewApi(tx)
				api2.SetupAll()
				req := httptest.NewRequest("GET", "/api/salary", nil)
				req.Header.Set("Authorization", "Bearer "+respBody["token"].(string))
				resp, _ := api2.Router.Test(req)
				a.Equal(200, resp.StatusCode)
			} else {
				a.Equal(d.expectedBody, respBody)
			}
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	api := testhelper.NewTestApi(tx)

	// Create a user first
	resp := api.Test(http.MethodPost, "/api/users", util.M{
		"email":    "test@example.com",
		"password": "password123",
		"salary":   5000.00,
	})
	a.Equal(201, resp.StatusCode)

	data := []struct {
		name           string
		body           util.M
		expectedStatus int
		expectedBody   util.M
	}{
		{
			"login successfully",
			util.M{
				"email":    "test@example.com",
				"password": "password123",
			},
			201,
			nil,
		},
		{
			"invalid credentials - user not found",
			util.M{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			401,
			util.M{"error": "invalid email or password"},
		},
		{
			"invalid credentials - wrong password",
			util.M{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			401,
			util.M{"error": "invalid email or password"},
		},
		{
			"missing required fields",
			util.M{},
			400,
			util.M{"errors": util.M{
				"email":    []any{"is required"},
				"password": []any{"is required"},
			}},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			a := assert.New(t)

			var respBody util.M

			resp := api.Test(http.MethodPost, "/api/sessions", d.body)
			api.UnmarshalBody(resp.Body, &respBody)
			a.Equal(d.expectedStatus, resp.StatusCode)

			if d.expectedStatus == 201 {
				a.NotEmpty(respBody["token"])

				// Verify the token works
				api2 := originalApi.NewApi(tx)
				api2.SetupAll()
				req := httptest.NewRequest("GET", "/api/salary", nil)
				req.Header.Set("Authorization", "Bearer "+respBody["token"].(string))
				tokenResp, _ := api2.Router.Test(req)
				a.Equal(200, tokenResp.StatusCode)
			} else {
				a.Equal(d.expectedBody, respBody)
			}
		})
	}
}

func TestApi_ValidateTokenMiddleware(t *testing.T) {
	t.Parallel()

	tx := testhelper.NewTestPostgresTx(t)
	api := originalApi.NewApi(tx)
	api.Router.Use(api.ValidateTokenMiddleware())

	userID := uuid.New()

	tests := []struct {
		name           string
		setupToken     func() string
		handler        fiber.Handler
		expectedStatus int
		expectedError  string
	}{
		{
			name: "with valid token continue to the next handler",
			setupToken: func() string {
				return domain.CreateToken(userID, time.Minute)
			},
			handler: func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusNoContent)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "expired token should return unauthorized",
			setupToken: func() string {
				return domain.CreateToken(userID, -time.Minute)
			},
			handler: func(c *fiber.Ctx) error {
				panic("cannot get here")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid or expired JWT",
		},
		{
			name: "malformed token should return unauthorized",
			setupToken: func() string {
				return "invalid-token"
			},
			handler: func(c *fiber.Ctx) error {
				panic("cannot get here")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid or expired JWT",
		},
		{
			name:           "empty token should return unauthorized",
			setupToken:     func() string { return "" },
			expectedStatus: http.StatusBadRequest,
			expectedError:  "missing or malformed JWT",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			token := tt.setupToken()

			path := fmt.Sprintf("/test-%d", i)
			api.Router.Get(path, tt.handler)

			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := api.Router.Test(req)
			a.NoError(err)
			a.Equal(tt.expectedStatus, resp.StatusCode)

			if tt.expectedError != "" {
				var respBody util.M
				err := json.NewDecoder(resp.Body).Decode(&respBody)
				a.NoError(err)
				a.Equal(tt.expectedError, respBody["error"])
			}
		})
	}
}

func TestApi_PutUserIDMiddleware(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	tx := testhelper.NewTestPostgresTx(t)
	api := originalApi.NewApi(tx)
	middleware := api.PutUserIDMiddleware()

	userID := uuid.New()
	token := domain.CreateToken(userID, time.Minute)

	api.Router.Use(api.ValidateTokenMiddleware())
	api.Router.Use(middleware)
	api.Router.Get("/test", func(c *fiber.Ctx) error {
		a.Equal(userID.String(), c.Locals("user_id"))
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := api.Router.Test(req)
	a.NoError(err)
	a.Equal(http.StatusOK, resp.StatusCode)
}
