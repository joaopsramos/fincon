package handler_test

import (
	"net/http"
	"testing"

	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/joaopsramos/fincon/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestSalaryHandler_Get(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	f := testhelper.NewFactory(tx)
	user := f.InsertUser()
	anotherUser := f.InsertUser()

	api := testhelper.NewTestApi(user.ID, tx)
	anotherUserApi := testhelper.NewTestApi(anotherUser.ID, tx)

	salaries := []*domain.Salary{
		{Amount: 50000, UserID: user.ID},
		{Amount: 100000, UserID: anotherUser.ID},
	}
	f.InsertSalary(salaries...)

	data := []struct {
		name     string
		api      *testhelper.TestApi
		status   int
		expected util.M
	}{
		{
			"get user salary",
			api,
			200,
			util.M{"amount": float64(salaries[0].Amount)},
		},
		{
			"ensure user only gets his salary",
			anotherUserApi,
			200,
			util.M{"amount": float64(salaries[1].Amount)},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			var respBody util.M

			resp := d.api.Test(http.MethodGet, "/api/salary")
			d.api.UnmarshalBody(resp.Body, &respBody)
			assert.Equal(resp.StatusCode, d.status)
			assert.Equal(d.expected, respBody)
		})
	}
}
