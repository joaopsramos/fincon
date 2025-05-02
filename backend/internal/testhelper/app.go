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
	"github.com/joaopsramos/fincon/internal/auth"
	"gorm.io/gorm"
)

type TestApp struct {
	app   *api.App
	token string
}

func NewTestApp(tx *gorm.DB, userID ...uuid.UUID) *TestApp {
	app := api.NewApp(tx)
	app.SetupAll()
	var token string

	if len(userID) > 0 {
		token = auth.GenerateJWTToken(userID[0], time.Minute*1)
	}

	return &TestApp{app: app, token: token}
}

func (t *TestApp) Test(method string, path string, body ...any) *http.Response {
	var bodyReader io.Reader

	if len(body) > 0 {
		encodedBody, _ := json.Marshal(body[0])
		bodyReader = bytes.NewReader(encodedBody)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	if t.token != "" {
		req.Header.Set("Authorization", "Bearer "+t.token)
	}

	w := httptest.NewRecorder()
	t.app.Router.ServeHTTP(w, req)

	return w.Result()
}

func (t *TestApp) UnmarshalBody(body io.ReadCloser, dst any) {
	err := json.NewDecoder(body).Decode(dst)
	if err != nil && !errors.Is(err, io.EOF) {
		panic(err)
	}
}
