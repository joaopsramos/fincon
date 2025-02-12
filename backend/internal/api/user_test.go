package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	originalApi "github.com/joaopsramos/fincon/internal/api"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/joaopsramos/fincon/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestUserHandler_Create(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
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
				"salary": util.M{"amount": float64(5000), "currency": "BRL"},
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
			var respBody util.M

			resp := api.Test(http.MethodPost, "/api/users", d.body)
			api.UnmarshalBody(resp.Body, &respBody)
			assert.Equal(d.expectedStatus, resp.StatusCode)

			if d.expectedStatus == 201 {
				assert.Equal(d.expectedBody["email"], respBody["email"])
				assert.Equal(d.expectedBody["salary"], respBody["salary"])
				assert.NotEmpty(respBody["user"].(util.M)["id"])
				assert.NotEmpty(respBody["token"])

				// assert valid token
				api2 := originalApi.NewApi(tx)
				api2.SetupAll()
				req := httptest.NewRequest("GET", "/api/salary", nil)
				req.Header.Set("Authorization", "Bearer "+respBody["token"].(string))
				resp, _ := api2.Router.Test(req)
				assert.Equal(200, resp.StatusCode)
			} else {
				assert.Equal(d.expectedBody, respBody)
			}
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	api := testhelper.NewTestApi(tx)

	// Create a user first
	resp := api.Test(http.MethodPost, "/api/users", util.M{
		"email":    "test@example.com",
		"password": "password123",
		"salary":   5000.00,
	})
	assert.Equal(201, resp.StatusCode)

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
			var respBody util.M

			resp := api.Test(http.MethodPost, "/api/sessions", d.body)
			api.UnmarshalBody(resp.Body, &respBody)
			assert.Equal(d.expectedStatus, resp.StatusCode)

			if d.expectedStatus == 201 {
				assert.NotEmpty(respBody["token"])

				// Verify the token works
				api2 := originalApi.NewApi(tx)
				api2.SetupAll()
				req := httptest.NewRequest("GET", "/api/salary", nil)
				req.Header.Set("Authorization", "Bearer "+respBody["token"].(string))
				tokenResp, _ := api2.Router.Test(req)
				assert.Equal(200, tokenResp.StatusCode)
			} else {
				assert.Equal(d.expectedBody, respBody)
			}
		})
	}
}
