package api_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	z "github.com/Oudwins/zog"
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

	api.Router.Get("/not-found", func(w http.ResponseWriter, r *http.Request) {
		api.HandleError(w, errs.NewNotFound("some resource"))
	})

	api.Router.Get("/some-error", func(w http.ResponseWriter, r *http.Request) {
		api.HandleError(w, errors.New("some error"))
	})

	data := []struct {
		name           string
		url            string
		expectedStatus int
		expectedBody   util.M
	}{
		{"not found error", "/not-found", 404, util.M{"error": "some resource not found"}},
		{"any other error", "/some-error", 500, nil},
	}

	for _, d := range data {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, d.url, nil)
		api.Router.ServeHTTP(w, req)

		var respBody util.M
		_ = json.NewDecoder(w.Body).Decode(&respBody)

		assert.Equal(d.expectedStatus, w.Code)
		assert.Equal(d.expectedBody, respBody)
	}
}

func Test_HandleZodError(t *testing.T) {
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	api := api.NewApi(tx)
	api.SetupMiddlewares()

	api.Router.Post("/some-route", func(w http.ResponseWriter, r *http.Request) {
		schema := z.Struct(z.Schema{
			"name": z.String().Required(),
		})

		var dst struct{ Name string }
		errs := util.ParseZodSchema(schema, r.Body, &dst)
		api.HandleZodError(w, errs)
	})

	req := httptest.NewRequest(http.MethodPost, "/some-route", strings.NewReader(`{"not-name": 1}`))
	w := httptest.NewRecorder()
	api.Router.ServeHTTP(w, req)

	var respBody util.M
	_ = json.NewDecoder(w.Body).Decode(&respBody)

	assert.Equal(400, w.Code)
	assert.Equal(util.M{"errors": util.M{"name": []any{"is required"}}}, respBody)
}

func Test_InvalidJSONBody(t *testing.T) {
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	api := api.NewApi(tx)
	api.SetupMiddlewares()

	api.Router.Post("/some-route", func(w http.ResponseWriter, r *http.Request) {
		var body util.M
		err := json.NewDecoder(r.Body).Decode(&body)
		api.InvalidJSONBody(w, err)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/some-route", strings.NewReader(`{"invalid": json}`))
	api.Router.ServeHTTP(w, req)

	var respBody util.M
	_ = json.NewDecoder(w.Body).Decode(&respBody)

	assert.Equal(400, w.Code)
	assert.Equal(util.M{"error": "invalid json body"}, respBody)
}
