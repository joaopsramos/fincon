package testhelper

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/api"
	"github.com/joaopsramos/fincon/internal/auth"
	"github.com/joaopsramos/fincon/internal/mail"
	"gorm.io/gorm"
)

type TestApp struct {
	*api.App
	token  string
	Mailer mail.Mailer
}

type TestAppOpts struct {
	UserID       uuid.UUID
	Logger       *slog.Logger
	Mailer       mail.Mailer
	WithoutSetup bool
}

func NewTestApp(t testing.TB, tx *gorm.DB, options ...TestAppOpts) *TestApp {
	var opts TestAppOpts
	if len(options) > 0 {
		opts = options[0]
	}

	if opts.Mailer == nil {
		opts.Mailer = NewMockMailer(t)
	}

	app := api.NewApp(tx, opts.Logger, opts.Mailer)

	if !opts.WithoutSetup {
		app.SetupAll()
	}

	var token string
	if opts.UserID != uuid.Nil {
		token = auth.GenerateJWTToken(opts.UserID, time.Minute)
	}

	return &TestApp{
		App:    app,
		token:  token,
		Mailer: opts.Mailer,
	}
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
	t.Router.ServeHTTP(w, req)

	return w.Result()
}

func (t *TestApp) UnmarshalBody(body io.ReadCloser, dst any) {
	err := json.NewDecoder(body).Decode(dst)
	if err != nil && !errors.Is(err, io.EOF) {
		panic(err)
	}
}
