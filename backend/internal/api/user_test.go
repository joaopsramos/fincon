package api_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/mail"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/joaopsramos/fincon/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
			app := testhelper.NewTestApp(t, tx)

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

			// assert valid token without using helper function
			req := httptest.NewRequest("GET", "/api/salary", nil)
			req.Header.Set("Authorization", "Bearer "+respBody["token"].(string))

			w := httptest.NewRecorder()
			app.Router.ServeHTTP(w, req)

			resp2 := w.Result()
			a.Equal(200, resp2.StatusCode)
		})
	}
}

func TestUserHandler_UserLogin(t *testing.T) {
	t.Parallel()
	tx := testhelper.NewTestPostgresTx(t)
	app := testhelper.NewTestApp(t, tx)

	// Create a user first
	resp := app.Test(http.MethodPost, "/api/users", util.M{
		"email":    "test@example.com",
		"password": "password123",
		"salary":   5000.00,
	})
	assert.Equal(t, 201, resp.StatusCode)

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
			assert := assert.New(t)

			var respBody util.M

			resp := app.Test(http.MethodPost, "/api/sessions", d.body)
			app.UnmarshalBody(resp.Body, &respBody)
			assert.Equal(d.expectedStatus, resp.StatusCode)

			if d.expectedStatus == 201 {
				assert.NotEmpty(respBody["token"])

				// Verify the token works
				app2 := testhelper.NewTestApp(t, tx)
				req := httptest.NewRequest("GET", "/api/salary", nil)
				req.Header.Set("Authorization", "Bearer "+respBody["token"].(string))
				w := httptest.NewRecorder()
				app2.Router.ServeHTTP(w, req)
				resp := w.Result()
				assert.Equal(200, resp.StatusCode)
			} else {
				assert.Equal(d.expectedBody, respBody)
			}
		})
	}
}

func TestUserHandler_ForgotPassword(t *testing.T) {
	t.Parallel()
	tx := testhelper.NewTestPostgresTx(t)
	mailer := testhelper.NewMockMailer(t)
	app := testhelper.NewTestApp(t, tx, testhelper.TestAppOpts{Mailer: mailer})

	f := testhelper.NewFactory(tx)
	user := f.InsertUser()

	tests := []struct {
		name           string
		body           any
		expectedStatus int
		expectedBody   util.M
		setupMocks     func()
	}{
		{
			name: "successful forgot password request",
			body: util.M{
				"email": user.Email,
			},
			expectedStatus: 200,
			expectedBody:   nil,
			setupMocks: func() {
				mailer.EXPECT().Send(mock.AnythingOfType("mail.Email")).Return(nil).Once()
			},
		},
		{
			name: "user not found",
			body: util.M{
				"email": "nonexistent@example.com",
			},
			expectedStatus: 404,
			expectedBody:   util.M{"error": "user not found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			resp := app.Test(http.MethodPost, "/api/password/forgot", tt.body)

			var respBody util.M
			app.UnmarshalBody(resp.Body, &respBody)
			assert.Equal(tt.expectedBody, respBody)

			assert.Equal(tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestUserHandler_ResetPassword(t *testing.T) {
	t.Parallel()
	tx := testhelper.NewTestPostgresTx(t)
	mailer := testhelper.NewMockMailer(t)
	app := testhelper.NewTestApp(t, tx, testhelper.TestAppOpts{Mailer: mailer})

	factory := testhelper.NewFactory(tx)
	user := factory.InsertUser()

	tests := []struct {
		name           string
		setupToken     func(t *testing.T) string
		password       string
		expectedStatus int
		expectedBody   util.M
		setupMocks     func()
	}{
		{
			name:     "successful password reset request",
			password: "newPassword123",
			setupToken: func(t *testing.T) string {
				var token string
				mailer.EXPECT().Send(mock.AnythingOfType("mail.Email")).Run(func(email mail.Email) {
					url, err := url.Parse(email.Data["Link"].(string))
					require.NoError(t, err)

					token = url.Query().Get("token")
				}).Return(nil).Once()

				app.Test(http.MethodPost, "/api/password/forgot", util.M{"email": user.Email})

				return token
			},
			expectedStatus: 200,
			expectedBody:   nil,
		},
		{
			name:           "password too short",
			password:       "short",
			setupToken:     func(t *testing.T) string { return uuid.NewString() },
			expectedStatus: 400,
			expectedBody:   util.M{"errors": util.M{"password": []any{"must contain at least 8 characters"}}},
		},
		{
			name:           "password too long",
			password:       "ThisIsAReallyLongPasswordThatExceedsTheMaximumLengthOf72CharactersAndShouldFailValidation",
			setupToken:     func(t *testing.T) string { return uuid.NewString() },
			expectedStatus: 400,
			expectedBody:   util.M{"errors": util.M{"password": []any{"must have at most 72 characters"}}},
		},
		{
			name:           "no password",
			password:       "",
			setupToken:     func(t *testing.T) string { return uuid.NewString() },
			expectedStatus: 400,
			expectedBody:   util.M{"errors": util.M{"password": []any{"is required"}}},
		},
		{
			name:     "user not found",
			password: "newPassword123",
			setupToken: func(t *testing.T) string {
				return uuid.NewString()
			},
			expectedStatus: 404,
			expectedBody:   util.M{"error": "user not found"},
		},
		{
			name:     "expired token",
			password: "newPassword123",
			setupToken: func(t *testing.T) string {
				return factory.InsertUserToken(&domain.UserToken{
					UserID:    user.ID,
					ExpiresAt: time.Now().UTC().Add(-24 * time.Hour),
				}).Token.String()
			},
			expectedStatus: 400,
			expectedBody:   util.M{"error": "invalid or expired token"},
		},
		{
			name:     "used token",
			password: "newPassword123",
			setupToken: func(t *testing.T) string {
				return factory.InsertUserToken(&domain.UserToken{
					UserID: user.ID,
					Used:   true,
				}).Token.String()
			},
			expectedStatus: 400,
			expectedBody:   util.M{"error": "invalid or expired token"},
		},
		{
			name:     "invalid token",
			password: "newPassword123",
			setupToken: func(t *testing.T) string {
				return "invalid-token"
			},
			expectedStatus: 400,
			expectedBody:   util.M{"error": "invalid or expired token"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			token := tt.setupToken(t)
			newPassword := tt.password

			resp := app.Test(http.MethodPost, "/api/password/reset", util.M{"token": token, "password": newPassword})

			var respBody util.M
			app.UnmarshalBody(resp.Body, &respBody)

			assert.Equal(tt.expectedBody, respBody)
			assert.Equal(tt.expectedStatus, resp.StatusCode)
		})
	}
}
