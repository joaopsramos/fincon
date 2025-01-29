package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/stretchr/testify/assert"
)

func TestExpenseController_Create(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestSQLiteTx(t)
	api := testhelper.NewTestApi(tx)
	f := testhelper.NewFactory(tx)

	var respBody map[string]any

	data := []struct {
		name     string
		body     map[string]any
		status   int
		expected map[string]any
	}{
		{
			"ensure required fields",
			map[string]any{},
			400,
			map[string]any{"errors": map[string]any{
				"name":    []any{"is required"},
				"value":   []any{"is required"},
				"date":    []any{"is required"},
				"goal_id": []any{"is required"},
			}},
		},
		{
			"invalid date",
			map[string]any{"name": "Food", "value": 123.45, "date": "2025-13-25", "goal_id": 1},
			400,
			map[string]any{"errors": map[string]any{"date": []any{"time is invalid"}}},
		},
		{
			"goal not found",
			map[string]any{"name": "Food", "value": 123.45, "date": "2025-12-15", "goal_id": 10},
			400,
			map[string]any{"error": "goal not found"},
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
		map[string]any{"name": "Food", "value": 123.45, "date": "2025-01-15", "goal_id": 1},
	)
	json.NewDecoder(resp.Body).Decode(&respBody)

	assert.Equal(resp.StatusCode, 201)
	assert.Equal("Food", respBody["name"])
	assert.Equal(map[string]any{"amount": 123.45, "currency": "BRL"}, respBody["value"])
	assert.Equal("2025-01-15T00:00:00Z", respBody["date"])
	assert.Equal(1.0, respBody["goal_id"])
	assert.Contains(respBody, "id")

	expectedKeys := map[string]struct{}{
		"id":      {},
		"name":    {},
		"value":   {},
		"date":    {},
		"goal_id": {},
	}

	for k := range respBody {
		assert.Contains(expectedKeys, k)
	}
}
