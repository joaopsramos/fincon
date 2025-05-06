package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/joaopsramos/fincon/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestUserHandler_CreateUser(t *testing.T) {
	t.Parallel()
	tx := testhelper.NewTestPostgresTx(t)

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
			app := testhelper.NewTestApp(tx)
			app2 := testhelper.NewTestApp(tx)

			var respBody util.M

			resp := app.Test(http.MethodPost, "/api/users", d.body)
			app.UnmarshalBody(resp.Body, &respBody)
			a.Equal(d.expectedStatus, resp.StatusCode)

			if d.expectedStatus != 201 {
				a.Equal(d.expectedBody, respBody)
				return
			}

			a.Equal(d.expectedBody["email"], respBody["email"])
			a.Equal(d.expectedBody["salary"], respBody["salary"])
			a.NotEmpty(respBody["user"].(util.M)["id"])
			a.NotEmpty(respBody["token"])

			// assert valid token
			req := httptest.NewRequest("GET", "/api/salary", nil)
			req.Header.Set("Authorization", "Bearer "+respBody["token"].(string))

			w := httptest.NewRecorder()
			app2.Router.ServeHTTP(w, req)

			resp2 := w.Result()
			a.Equal(200, resp2.StatusCode)
		})
	}
}

func TestUserHandler_UserLogin(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	app := testhelper.NewTestApp(tx)

	// Create a user first
	resp := app.Test(http.MethodPost, "/api/users", util.M{
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

			resp := app.Test(http.MethodPost, "/api/sessions", d.body)
			app.UnmarshalBody(resp.Body, &respBody)
			a.Equal(d.expectedStatus, resp.StatusCode)

			if d.expectedStatus == 201 {
				a.NotEmpty(respBody["token"])

				// Verify the token works
				app2 := testhelper.NewTestApp(tx)
				req := httptest.NewRequest("GET", "/api/salary", nil)
				req.Header.Set("Authorization", "Bearer "+respBody["token"].(string))
				w := httptest.NewRecorder()
				app2.Router.ServeHTTP(w, req)
				resp := w.Result()
				a.Equal(200, resp.StatusCode)
			} else {
				a.Equal(d.expectedBody, respBody)
			}
		})
	}
}
