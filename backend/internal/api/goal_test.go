package api_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/testhelper"
	"github.com/joaopsramos/fincon/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestGoalHandler_Index(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	f := testhelper.NewFactory(tx)
	user := f.InsertUser()
	anotherUser := f.InsertUser()

	api := testhelper.NewTestApi(user.ID, tx)
	anotherUserApi := testhelper.NewTestApi(anotherUser.ID, tx)

	goals := []*domain.Goal{
		{Name: "Comfort", Percentage: 40, UserID: user.ID},
		{Name: "Goals", Percentage: 20, UserID: user.ID},
		{Name: "Fixed costs", Percentage: 30, UserID: user.ID},
		{Name: "Pleasures", Percentage: 100, UserID: anotherUser.ID},
	}
	f.InsertGoal(goals...)

	data := []struct {
		name     string
		api      *testhelper.TestApi
		status   int
		expected []util.M
	}{
		{
			"get all goals",
			api,
			200,
			[]util.M{
				{"id": float64(goals[0].ID), "name": "Comfort", "percentage": float64(40)},
				{"id": float64(goals[1].ID), "name": "Goals", "percentage": float64(20)},
				{"id": float64(goals[2].ID), "name": "Fixed costs", "percentage": float64(30)},
			},
		},
		{
			"only return goals from the current user",
			anotherUserApi,
			200,
			[]util.M{
				{"id": float64(goals[3].ID), "name": "Pleasures", "percentage": float64(100)},
			},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			var respBody []util.M

			resp := d.api.Test(http.MethodGet, "/api/goals")
			d.api.UnmarshalBody(resp.Body, &respBody)
			assert.Equal(resp.StatusCode, d.status)
			assert.ElementsMatch(d.expected, respBody)
		})
	}
}

func TestGoalHandler_GetExpenses(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	f := testhelper.NewFactory(tx)
	user := f.InsertUser()
	anotherUser := f.InsertUser()

	api := testhelper.NewTestApi(user.ID, tx)
	anotherUserApi := testhelper.NewTestApi(anotherUser.ID, tx)

	goals := []*domain.Goal{
		{Name: "Comfort", Percentage: 40, UserID: user.ID},
		{Name: "Goals", Percentage: 60, UserID: user.ID},
		{Name: "Pleasures", Percentage: 100, UserID: anotherUser.ID},
	}
	f.InsertGoal(goals...)

	now := testhelper.MiddleOfMonth()

	expenses := []*domain.Expense{
		{Name: "Cake", Value: 123, Date: now, GoalID: goals[0].ID, UserID: user.ID},
		{Name: "Health", Value: 321, Date: now, GoalID: goals[0].ID, UserID: user.ID},
		{Name: "Mouse", Value: 49312, Date: now.AddDate(0, -1, 0), GoalID: goals[0].ID, UserID: user.ID},
		{Name: "Game", Value: 6000, Date: now.AddDate(-1, -1, 0), GoalID: goals[0].ID, UserID: user.ID},
		{Name: "Phone", Value: 333, Date: now, GoalID: goals[1].ID, UserID: user.ID},
		{Name: "PC", Value: 222, Date: now, GoalID: goals[2].ID, UserID: anotherUser.ID},
	}
	f.InsertExpense(expenses...)

	oneMonthAgo := now.AddDate(0, -1, 0)
	oneYearAndOneMonthAgo := now.AddDate(-1, -1, 0)

	data := []struct {
		name     string
		api      *testhelper.TestApi
		goalID   uint
		date     *time.Time
		status   int
		expected []util.M
	}{
		{
			"get all goal expenses of current month", api, goals[0].ID, nil, 200, []util.M{
				testhelper.FormatExpense(*expenses[0], *goals[0]),
				testhelper.FormatExpense(*expenses[1], *goals[0]),
			},
		},
		{
			"get expenses of a specific month", api, goals[0].ID, &oneMonthAgo, 200, []util.M{
				testhelper.FormatExpense(*expenses[2], *goals[0]),
			},
		},
		{
			"get expenses of a specific year and month", api, goals[0].ID, &oneYearAndOneMonthAgo, 200, []util.M{
				testhelper.FormatExpense(*expenses[3], *goals[0]),
			},
		},
		{
			"get expenses of another goal", api, goals[1].ID, &now, 200, []util.M{
				testhelper.FormatExpense(*expenses[4], *goals[1]),
			},
		},
		{
			"only return expenses from the current user", anotherUserApi, goals[2].ID, &now, 200, []util.M{
				testhelper.FormatExpense(*expenses[5], *goals[2]),
			},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			var respBody []util.M
			var query string

			if d.date != nil {
				query = fmt.Sprintf("?year=%d&month=%d", d.date.Year(), d.date.Month())
			}

			resp := d.api.Test(http.MethodGet, fmt.Sprintf("/api/goals/%d/expenses/%s", d.goalID, query))
			d.api.UnmarshalBody(resp.Body, &respBody)

			assert.Equal(d.status, resp.StatusCode)
			assert.ElementsMatch(d.expected, respBody)
		})
	}

	stringGoalID := fmt.Sprintf("%d", goals[0].ID)

	data2 := []struct {
		name     string
		goalID   string
		year     string
		month    string
		status   int
		expected util.M
	}{
		{"invalid year", stringGoalID, "invalid", "5", 400, util.M{"error": "invalid year"}},
		{"year < 1", stringGoalID, "0", "5", 400, util.M{"error": "invalid year"}},
		{"invalid month", stringGoalID, "2024", "invalid", 400, util.M{"error": "invalid month"}},
		{"month < 1", stringGoalID, "2024", "0", 400, util.M{"error": "invalid month"}},
		{"month > 12", stringGoalID, "2024", "13", 400, util.M{"error": "invalid month"}},
		{"invalid goal id", "invalid", "2024", "1", 400, util.M{"error": "invalid goal id"}},
	}

	for _, d := range data2 {
		t.Run(d.name, func(t *testing.T) {
			var respBody util.M

			resp := api.Test(http.MethodGet, fmt.Sprintf("/api/goals/%s/expenses?year=%s&month=%s", d.goalID, d.year, d.month))
			api.UnmarshalBody(resp.Body, &respBody)

			assert.Equal(d.status, resp.StatusCode)
			assert.Equal(d.expected, respBody)
		})
	}
}
