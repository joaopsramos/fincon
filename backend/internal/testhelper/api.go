package testhelper

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/api"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/util"
	"gorm.io/gorm"
)

type TestApi struct {
	api   *api.Api
	token string
}

func NewTestApi(userID uuid.UUID, tx *gorm.DB) *TestApi {
	api := api.NewApi(tx)
	api.Setup()

	return &TestApi{api: api, token: domain.CreateToken(userID, time.Minute*1)}
}

func (t *TestApi) Test(method string, path string, body util.M) *http.Response {
	var bodyReader io.Reader

	if body != nil {
		encodedBody, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(encodedBody)
	}

	req := httptest.NewRequest("POST", "/api/expenses", bodyReader)
	req.Header.Set("Authorization", "Bearer "+t.token)

	resp, _ := t.api.Router.Test(req)
	return resp
}
