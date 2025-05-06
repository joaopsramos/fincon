package api_test

import (
	"net/http"
	"testing"

	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/joaopsramos/fincon/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestSalaryHandler_GetSalary(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	f := testhelper.NewFactory(tx)
	user := f.InsertUser()
	anotherUser := f.InsertUser()

	app := testhelper.NewTestApp(t, tx, testhelper.TestAppOpts{UserID: user.ID})
	anotherUserApp := testhelper.NewTestApp(t, tx, testhelper.TestAppOpts{UserID: anotherUser.ID})

	salaries := []*domain.Salary{
		{Amount: 50000, UserID: user.ID},
		{Amount: 100000, UserID: anotherUser.ID},
	}
	f.InsertSalary(salaries...)

	data := []struct {
		name     string
		app      *testhelper.TestApp
		status   int
		expected util.M
	}{
		{
			"get user salary",
			app,
			200,
			util.M{"amount": 500.0},
		},
		{
			"ensure user only gets his salary",
			anotherUserApp,
			200,
			util.M{"amount": 1000.0},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			var respBody util.M

			resp := d.app.Test(http.MethodGet, "/api/salary")
			d.app.UnmarshalBody(resp.Body, &respBody)
			assert.Equal(resp.StatusCode, d.status)
			assert.Equal(d.expected, respBody)
		})
	}
}

func TestSalaryHandler_UpdateSalary(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	f := testhelper.NewFactory(tx)
	user := f.InsertUser()
	anotherUser := f.InsertUser()

	app := testhelper.NewTestApp(t, tx, testhelper.TestAppOpts{UserID: user.ID})
	anotherUserApp := testhelper.NewTestApp(t, tx, testhelper.TestAppOpts{UserID: anotherUser.ID})

	f.InsertSalary([]*domain.Salary{
		{Amount: 50000, UserID: user.ID},
		{Amount: 100000, UserID: anotherUser.ID},
	}...)

	data := []struct {
		name           string
		app            *testhelper.TestApp
		body           util.M
		expectedStatus int
		expectedBody   util.M
	}{
		{
			"update user salary",
			app,
			util.M{"amount": 1000},
			200,
			util.M{"amount": 1000.0},
		},
		{
			"ensure user only update his salary",
			anotherUserApp,
			util.M{"amount": 2000.50},
			200,
			util.M{"amount": 2000.50},
		},
		{
			"salary amount is required",
			app,
			util.M{"not-amount": 1},
			400,
			util.M{"errors": util.M{"amount": []any{"is required"}}},
		},
		{
			"salary amount must be greater than zero",
			app,
			util.M{"amount": -1},
			400,
			util.M{"errors": util.M{"amount": []any{"must be greater than 0"}}},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			var respBody util.M

			resp := d.app.Test(http.MethodPatch, "/api/salary", d.body)
			d.app.UnmarshalBody(resp.Body, &respBody)
			assert.Equal(resp.StatusCode, d.expectedStatus)
			assert.Equal(d.expectedBody, respBody)
		})
	}
}
