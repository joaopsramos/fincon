package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/joaopsramos/fincon/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestExpenseController_Create(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestSQLiteTx(t)
	api := testhelper.NewTestApi(tx)
	f := testhelper.NewFactory(tx)

	var respBody util.M

	data := []struct {
		name     string
		body     util.M
		status   int
		expected util.M
	}{
		{
			"ensure required fields",
			util.M{},
			400,
			util.M{"errors": util.M{
				"name":    []any{"is required"},
				"value":   []any{"is required"},
				"date":    []any{"is required"},
				"goal_id": []any{"is required"},
			}},
		},
		{
			"invalid date",
			util.M{"name": "Food", "value": 123.45, "date": "2025-13-25", "goal_id": 1},
			400,
			util.M{"errors": util.M{"date": []any{"time is invalid"}}},
		},
		{
			"goal not found",
			util.M{"name": "Food", "value": 123.45, "date": "2025-12-15", "goal_id": 10},
			400,
			util.M{"error": "goal not found"},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			resp := api.Test(http.MethodPost, "/api/expenses", d.body)
			json.NewDecoder(resp.Body).Decode(&respBody)
			assert.Equal(resp.StatusCode, 400)
			assert.Equal(d.expected, respBody)
			clear(respBody)
		})
	}

	f.InsertGoal(&domain.Goal{Name: "Comfort"})

	resp := api.Test(
		http.MethodPost,
		"/api/expenses",
		util.M{"name": "Food", "value": 123.45, "date": "2025-01-15", "goal_id": 1},
	)
	json.NewDecoder(resp.Body).Decode(&respBody)

	assert.Equal(resp.StatusCode, 201)

	assert.Contains(respBody, "id")
	delete(respBody, "id")

	assert.Equal(util.M{
		"name":    "Food",
		"value":   util.M{"amount": 123.45, "currency": "BRL"},
		"date":    "2025-01-15T00:00:00Z",
		"goal_id": 1.0,
	}, respBody)
}
