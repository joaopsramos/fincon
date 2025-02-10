package testhelper

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/api"
	"github.com/joaopsramos/fincon/internal/domain"
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

func (t *TestApi) Test(method string, path string, body ...any) *http.Response {
	var bodyReader io.Reader

	if len(body) > 0 {
		encodedBody, _ := json.Marshal(body[0])
		bodyReader = bytes.NewReader(encodedBody)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Authorization", "Bearer "+t.token)

	resp, err := t.api.Router.Test(req)
	if err != nil {
		panic(err)
	}

	return resp
}

func (t *TestApi) UnmarshalBody(body io.ReadCloser, dst any) {
	err := json.NewDecoder(body).Decode(dst)
	if err != nil && !errors.Is(err, io.EOF) {
		panic(err)
	}
}
