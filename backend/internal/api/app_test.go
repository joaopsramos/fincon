package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/api"
	"github.com/joaopsramos/fincon/internal/auth"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/joaopsramos/fincon/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestApp_ErrorHandler(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	app := testhelper.NewTestApp(t, tx)

	tests := []struct {
		name         string
		handler      http.HandlerFunc
		expectedCode int
		expectedBody util.M
	}{
		{
			name: "panic error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("test error")
			},
			expectedCode: 500,
			expectedBody: nil,
		},
		{
			name: "not found error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(util.M{"error": "resource not found"})
			},
			expectedCode: 404,
			expectedBody: util.M{"error": "resource not found"},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			routePath := fmt.Sprintf("/test-route-%d", i)
			app.Router.Get(routePath, tt.handler)

			req := httptest.NewRequest("GET", routePath, nil)
			w := httptest.NewRecorder()
			app.Router.ServeHTTP(w, req)

			var respBody util.M
			_ = json.NewDecoder(w.Body).Decode(&respBody)

			assert.Equal(tt.expectedCode, w.Code)
			assert.Equal(tt.expectedBody, respBody)
		})
	}
}

// TODO: Uncomment when rate limiter is better implemented

// func TestApp_GlobalRateLimiter(t *testing.T) {
// 	t.Parallel()
// 	assert := assert.New(t)
// 	tx := testhelper.NewTestPostgresTx(t)
// 	app := testhelper.NewTestApp(t, tx)
//
// 	app.Router.Get("/some-route", func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusNoContent)
// 	})
//
// 	for i := 1; i <= 101; i++ {
// 		req := httptest.NewRequest("GET", "/some-route", nil)
// 		w := httptest.NewRecorder()
// 		app.Router.ServeHTTP(w, req)
//
// 		if i == 101 {
// 			assert.Equal(http.StatusTooManyRequests, w.Code)
// 			retry := util.Must(strconv.Atoi(w.Header().Get("retry-after")))
// 			assert.InDelta(60, retry, 3)
// 			break
// 		}
//
// 		assert.Equal(http.StatusNoContent, w.Code)
// 	}
// }
//
// func TestApp_CreateUserRateLimiter(t *testing.T) {
// 	t.Parallel()
// 	assert := assert.New(t)
// 	tx := testhelper.NewTestPostgresTx(t)
// 	app := testhelper.NewTestApp(t, tx)
//
// 	for i := 1; i <= 6; i++ {
// 		user := util.M{"email": fmt.Sprintf("user-%d@mail.com", i), "password": 12345678, "salary": 1000}
// 		resp := app.Test(http.MethodPost, "/api/users", user)
//
// 		if i == 6 {
// 			assert.Equal(http.StatusTooManyRequests, resp.StatusCode)
// 			retry := util.Must(strconv.Atoi(resp.Header.Get("retry-after")))
// 			assert.InDelta(3600, retry, 3)
// 			break
// 		}
//
// 		assert.Equal(http.StatusCreated, resp.StatusCode)
// 	}
// }

// func TestApp_CreateSessionRateLimiter(t *testing.T) {
// 	t.Parallel()
// 	assert := assert.New(t)
// 	tx := testhelper.NewTestPostgresTx(t)
// 	app := testhelper.NewTestApp(t, tx)
//
// 	user := util.M{"email": "user@mail.com", "password": 12345678, "salary": 1000}
// 	resp := app.Test(http.MethodPost, "/api/users", user)
// 	assert.Equal(http.StatusCreated, resp.StatusCode)
//
// 	for i := 1; i <= 11; i++ {
// 		resp := app.Test(http.MethodPost, "/api/sessions", user)
//
// 		if i == 11 {
// 			assert.Equal(http.StatusTooManyRequests, resp.StatusCode)
// 			retry := util.Must(strconv.Atoi(resp.Header.Get("retry-after")))
// 			assert.InDelta(300, retry, 3)
// 			break
// 		}
//
// 		assert.Equal(http.StatusCreated, resp.StatusCode)
// 	}
// }

func TestApp_PutUserIDMiddleware(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	tx := testhelper.NewTestPostgresTx(t)
	app := testhelper.NewTestApp(t, tx, testhelper.TestAppOpts{WithoutSetup: true})

	userID := uuid.New()
	token := auth.GenerateJWTToken(userID, time.Minute)

	tokenAuth := auth.NewTokenAuth()
	app.Router.Use(jwtauth.Verifier(tokenAuth))
	app.Router.Use(jwtauth.Authenticator(tokenAuth))
	app.Router.Use(app.PutUserIDMiddleware)
	app.Router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		a.Equal(userID, r.Context().Value(api.UserIDKey).(uuid.UUID))
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	a.Equal(http.StatusOK, w.Result().StatusCode)
}
