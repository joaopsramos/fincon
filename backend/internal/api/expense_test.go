package api_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/repository"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/joaopsramos/fincon/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestExpenseHandler_Create(t *testing.T) {
	t.Parallel()
	tx := testhelper.NewTestPostgresTx(t)
	f := testhelper.NewFactory(tx)
	user := f.InsertUser()
	app := testhelper.NewTestApp(t, tx, testhelper.TestAppOpts{UserID: user.ID})

	goal := domain.Goal{Name: "Comfort", UserID: user.ID}
	f.InsertGoal(&goal)

	tests := []struct {
		name     string
		body     util.M
		status   int
		want     util.M
		validate func(t *testing.T, respBody util.M)
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
			nil,
		},
		{
			"invalid values",
			util.M{"name": "F", "value": 0.001, "date": "2025-13-25", "goal_id": 1},
			400,
			util.M{"errors": util.M{
				"name":  []any{"name must contain at least 2 characters"},
				"value": []any{"value must be greater than or equal to 0.01"},
				"date":  []any{"time is invalid"},
			}},
			nil,
		},
		{
			"goal not found",
			util.M{"name": "Food", "value": 123.45, "date": "2025-12-15", "goal_id": goal.ID + 1},
			404,
			util.M{"error": "goal not found"},
			nil,
		},
		{
			"successful expense creation",
			util.M{"name": "Food", "value": 123.45, "date": "2025-01-15", "goal_id": goal.ID},
			201,
			nil,
			func(t *testing.T, respBody util.M) {
				a := assert.New(t)

				data := respBody["data"].([]any)
				a.Len(data, 1)

				expense := data[0].(util.M)
				id := expense["id"].(float64)

				a.Equal(util.M{
					"id":      id,
					"name":    "Food",
					"value":   123.45,
					"date":    "2025-01-15T00:00:00Z",
					"goal_id": float64(goal.ID),
				}, expense)

				repo := repository.NewPostgresExpense(tx)
				_, err := repo.Get(context.Background(), uint(id), user.ID)
				a.Nil(err)

				_, err = repo.Get(context.Background(), uint(id), uuid.New())
				a.Equal("expense not found", err.Error())
			},
		},
		{
			"create expense with installments",
			util.M{"name": "Food", "value": 123.45, "date": "2025-01-15", "goal_id": goal.ID, "installments": 2},
			201,
			nil,
			func(t *testing.T, respBody util.M) {
				a := assert.New(t)

				data := respBody["data"].([]any)
				a.Len(data, 2)

				expense := data[0].(util.M)
				id := expense["id"].(float64)

				a.Equal(util.M{
					"id":      id,
					"name":    "Food (1/2)",
					"value":   123.45,
					"date":    "2025-01-15T00:00:00Z",
					"goal_id": float64(goal.ID),
				}, expense)

				repo := repository.NewPostgresExpense(tx)

				// assert it created two expenses, one for current month and another for the next one
				for i := range 2 {
					created, err := repo.AllByGoalID(context.Background(), goal.ID, 2025, time.Month(i+1), user.ID)
					a.Nil(err)
					idx := slices.IndexFunc(created, func(e domain.Expense) bool {
						return e.Name == fmt.Sprintf("Food (%d/2)", i+1)
					})
					a.NotEqual(idx, -1)
				}

				_, err := repo.Get(context.Background(), uint(id), uuid.New())
				a.Equal("expense not found", err.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			var respBody util.M

			resp := app.Test(http.MethodPost, "/api/expenses", tt.body)
			app.UnmarshalBody(resp.Body, &respBody)

			a.Equal(tt.status, resp.StatusCode)
			if tt.want != nil {
				a.Equal(tt.want, respBody)
			}

			if tt.validate != nil {
				tt.validate(t, respBody)
			}
		})
	}
}

func TestExpenseHandler_FindMatchingNames(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	f := testhelper.NewFactory(tx)
	user := f.InsertUser()
	app := testhelper.NewTestApp(t, tx, testhelper.TestAppOpts{UserID: user.ID})

	now := time.Now().UTC()
	// Use middle of month to avoid errors when subtracting/adding months
	now = time.Date(now.Year(), now.Month(), 15, 0, 0, 0, 0, time.UTC)

	goals := []*domain.Goal{{Name: "Comfort", UserID: user.ID}, {Name: "Goal", UserID: user.ID}}
	f.InsertGoal(goals...)

	f.InsertExpense([]*domain.Expense{
		{Name: "Apple", GoalID: goals[0].ID, UserID: user.ID, Date: now},
		{Name: "Application", GoalID: goals[1].ID, UserID: user.ID, Date: now.AddDate(0, -1, 0)},
		{Name: "Billó", GoalID: goals[1].ID, UserID: user.ID, Date: now},
		// Random user
		{Name: "Applic", GoalID: goals[1].ID, UserID: uuid.New(), Date: now},
	}...)

	var respBody []string

	data := []struct {
		query    string
		expected []string
	}{
		{"App", []string{"Apple", "Application"}},
		{"bill", []string{"Billó"}},
		{"pL", []string{"Apple", "Application"}},
		{"LL", []string{"Billó"}},
		{"lo", []string{"Billó"}},
		{"cat", []string{"Application"}},
		{"wal", []string{}},
	}

	for _, d := range data {
		resp := app.Test(http.MethodGet, "/api/expenses/matching-names?query="+d.query)
		app.UnmarshalBody(resp.Body, &respBody)

		assert.Equal(200, resp.StatusCode)
		assert.ElementsMatch(d.expected, respBody)

		clear(respBody)
	}

	var errBody util.M

	for _, q := range []string{"", "a"} {
		resp := app.Test(http.MethodGet, "/api/expenses/matching-names?query="+q)
		app.UnmarshalBody(resp.Body, &errBody)
		assert.Equal(400, resp.StatusCode)
		assert.Equal(util.M{"error": "query must be present and have at least 2 characters"}, errBody)

	}
}

func TestExpenseHandler_Update(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	f := testhelper.NewFactory(tx)
	user := f.InsertUser()
	app := testhelper.NewTestApp(t, tx, testhelper.TestAppOpts{UserID: user.ID})

	goal1 := domain.Goal{Name: "Comfort", UserID: user.ID}
	f.InsertGoal(&goal1)

	goal2 := domain.Goal{Name: "Pleasures", UserID: user.ID}
	f.InsertGoal(&goal2)

	expense := domain.Expense{Name: "Food", Value: 12345, Date: time.Now().UTC(), GoalID: goal1.ID, UserID: user.ID}
	f.InsertExpense(&expense)

	var respBody util.M

	data := []struct {
		name     string
		body     util.M
		status   int
		expected util.M
	}{
		{
			"invalid name",
			util.M{"name": "F"},
			400,
			util.M{"errors": util.M{"name": []any{"name must contain at least 2 characters"}}},
		},
		{
			"invalid value",
			util.M{"value": 0.001},
			400,
			util.M{"errors": util.M{"value": []any{"value must be greater than 0.01"}}},
		},
		{
			"invalid date",
			util.M{"name": "Food", "value": 123.45, "date": "2025-13-25"},
			400,
			util.M{"errors": util.M{"date": []any{"time is invalid"}}},
		},
		{
			"invalid goal",
			util.M{"goal_id": rand.Int()},
			404,
			util.M{"error": "goal not found"},
		},
		{
			"update all fields",
			util.M{"name": "Groceries", "value": 543.21, "date": "2023-01-15", "goal_id": goal2.ID},
			200,
			util.M{
				"id":      float64(expense.ID),
				"name":    "Groceries",
				"value":   543.21,
				"date":    "2023-01-15T00:00:00Z",
				"goal_id": float64(goal2.ID),
			},
		},
		{
			"update only name",
			util.M{"name": "Health"},
			200,
			util.M{
				"id":      float64(expense.ID),
				"name":    "Health",
				"value":   543.21,
				"date":    "2023-01-15T00:00:00Z",
				"goal_id": float64(goal2.ID),
			},
		},
		{
			"update only value",
			util.M{"value": 150.00},
			200,
			util.M{
				"id":      float64(expense.ID),
				"name":    "Health",
				"value":   150.00,
				"date":    "2023-01-15T00:00:00Z",
				"goal_id": float64(goal2.ID),
			},
		},
		{
			"update only date",
			util.M{"date": "2022-01-15"},
			200,
			util.M{
				"id":      float64(expense.ID),
				"name":    "Health",
				"value":   150.00,
				"date":    "2022-01-15T00:00:00Z",
				"goal_id": float64(goal2.ID),
			},
		},
		{
			"update only goal id",
			util.M{"goal_id": goal1.ID},
			200,
			util.M{
				"id":      float64(expense.ID),
				"name":    "Health",
				"value":   150.00,
				"date":    "2022-01-15T00:00:00Z",
				"goal_id": float64(goal1.ID),
			},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			resp := app.Test(http.MethodPatch, fmt.Sprintf("/api/expenses/%d", expense.ID), d.body)
			app.UnmarshalBody(resp.Body, &respBody)
			assert.Equal(d.status, resp.StatusCode)
			assert.Equal(d.expected, respBody)
			clear(respBody)
		})
	}

	resp := app.Test(http.MethodPatch, "/api/expenses/invalid-id", util.M{})
	app.UnmarshalBody(resp.Body, &respBody)
	assert.Equal(400, resp.StatusCode)
	assert.Equal(util.M{"error": "invalid expense id"}, respBody)

	anotherUserApp := testhelper.NewTestApp(t, tx, testhelper.TestAppOpts{UserID: uuid.New()})
	resp = anotherUserApp.Test(http.MethodPatch, fmt.Sprintf("/api/expenses/%d", expense.ID), util.M{})
	app.UnmarshalBody(resp.Body, &respBody)
	assert.Equal(404, resp.StatusCode)
	assert.Equal(util.M{"error": "expense not found"}, respBody)
}

func TestExpenseHandler_UpdateGoal(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	f := testhelper.NewFactory(tx)
	user := f.InsertUser()
	app := testhelper.NewTestApp(t, tx, testhelper.TestAppOpts{UserID: user.ID})

	goal1 := domain.Goal{Name: "Comfort", UserID: user.ID}
	goal2 := domain.Goal{Name: "Pleasures", UserID: user.ID}
	f.InsertGoal(&goal1)
	f.InsertGoal(&goal2)

	expense := domain.Expense{Value: 123, GoalID: goal1.ID, UserID: user.ID}
	f.InsertExpense(&expense)

	var respBody util.M

	data := []struct {
		name     string
		body     util.M
		status   int
		expected util.M
	}{
		{
			"goal not found",
			util.M{"goal_id": goal1.ID + 10},
			404,
			util.M{"error": "goal not found"},
		},
		{
			"update goal",
			util.M{"goal_id": goal2.ID},
			200,
			util.M{
				"id":      float64(expense.ID),
				"name":    expense.Name,
				"value":   1.23,
				"date":    testhelper.DateToJsonString(expense.Date),
				"goal_id": float64(goal2.ID),
			},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			resp := app.Test(http.MethodPatch, fmt.Sprintf("/api/expenses/%d/update-goal", expense.ID), d.body)
			app.UnmarshalBody(resp.Body, &respBody)
			assert.Equal(d.status, resp.StatusCode)
			assert.Equal(d.expected, respBody)
			clear(respBody)
		})
	}

	resp := app.Test(http.MethodPatch, "/api/expenses/invalid-id/update-goal", util.M{"goal_id": goal1.ID})
	app.UnmarshalBody(resp.Body, &respBody)
	assert.Equal(400, resp.StatusCode)
	assert.Equal(util.M{"error": "invalid expense id"}, respBody)

	anotherUserApp := testhelper.NewTestApp(t, tx, testhelper.TestAppOpts{UserID: uuid.New()})
	resp = anotherUserApp.Test(http.MethodPatch, fmt.Sprintf("/api/expenses/%d/update-goal", expense.ID), util.M{"goal_id": goal1.ID})
	app.UnmarshalBody(resp.Body, &respBody)
	assert.Equal(404, resp.StatusCode)
	assert.Equal(util.M{"error": "expense not found"}, respBody)
}

func TestExpenseHandler_Delete(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	f := testhelper.NewFactory(tx)
	user := f.InsertUser()
	app := testhelper.NewTestApp(t, tx, testhelper.TestAppOpts{UserID: user.ID})

	expense := domain.Expense{GoalID: f.InsertGoal().ID, UserID: user.ID}
	f.InsertExpense(&expense)

	var respBody util.M

	anotherUserApp := testhelper.NewTestApp(t, tx, testhelper.TestAppOpts{UserID: uuid.New()})

	data := []struct {
		name      string
		app       *testhelper.TestApp
		expenseID string
		status    int
		expected  util.M
	}{
		{
			"invalid expense id",
			app,
			"invalid-id",
			400,
			util.M{"error": "invalid expense id"},
		},
		{
			"only owner can delete",
			anotherUserApp,
			fmt.Sprintf("%d", expense.ID),
			404,
			util.M{"error": "expense not found"},
		},
		{
			"successfull delete",
			app,
			fmt.Sprintf("%d", expense.ID),
			204,
			util.M{},
		},
		{
			"not found after delete",
			app,
			fmt.Sprintf("%d", expense.ID),
			404,
			util.M{"error": "expense not found"},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			resp := d.app.Test(http.MethodDelete, "/api/expenses/"+d.expenseID)
			d.app.UnmarshalBody(resp.Body, &respBody)
			assert.Equal(d.status, resp.StatusCode)
			assert.Equal(d.expected, respBody)
			clear(respBody)
		})
	}

	repo := repository.NewPostgresExpense(tx)
	_, err := repo.Get(context.Background(), expense.ID, user.ID)
	assert.Equal("expense not found", err.Error())
}
