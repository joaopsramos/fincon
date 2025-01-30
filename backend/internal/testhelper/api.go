package testhelper

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/joaopsramos/fincon/internal/api"
	"gorm.io/gorm"
)

type TestApi struct {
	api *api.Api
}

func NewTestApi(tx *gorm.DB) *TestApi {
	api := api.NewApi(tx)
	api.Setup()

	return &TestApi{api: api}
}

func (t *TestApi) Test(method string, path string, body map[string]any) *http.Response {
	var bodyReader io.Reader

	if body != nil {
		encodedBody, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(encodedBody)
	}

	req := httptest.NewRequest("POST", "/api/expenses", bodyReader)
	resp, _ := t.api.Router.Test(req)
	return resp
}
