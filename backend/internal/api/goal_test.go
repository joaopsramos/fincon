package api_test

import (
	"fmt"
	"net/http"
	"reflect"
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

	api := testhelper.NewTestApi(tx, user.ID)
	anotherUserApi := testhelper.NewTestApi(tx, anotherUser.ID)

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

func TestGoalHandler_GetGoalExpenses(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	f := testhelper.NewFactory(tx)
	user := f.InsertUser()
	anotherUser := f.InsertUser()

	api := testhelper.NewTestApi(tx, user.ID)
	anotherUserApi := testhelper.NewTestApi(tx, anotherUser.ID)

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
				query = fmt.Sprintf("year=%d&month=%d", d.date.Year(), d.date.Month())
			}

			resp := d.api.Test(http.MethodGet, fmt.Sprintf("/api/goals/%d/expenses?%s", d.goalID, query))
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

func TestGoalHandler_UpdateGoals(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tx := testhelper.NewTestPostgresTx(t)
	f := testhelper.NewFactory(tx)
	user := f.InsertUser()
	anotherUser := f.InsertUser()

	api := testhelper.NewTestApi(tx, user.ID)
	anotherUserApi := testhelper.NewTestApi(tx, anotherUser.ID)
	_ = anotherUserApi

	defaultPercentages := domain.DefaultGoalPercentages()
	goals := make([]*domain.Goal, 0, len(defaultPercentages)*2)
	for _, u := range []domain.User{user, anotherUser} {
		for name, percentage := range defaultPercentages {
			goals = append(goals, &domain.Goal{Name: name, Percentage: percentage, UserID: u.ID})
		}
	}
	f.InsertGoal(goals...)

	formatGoal := func(g *domain.Goal, percentage uint) util.M {
		return util.M{"id": float64(g.ID), "name": string(g.Name), "percentage": float64(percentage)}
	}

	data := []struct {
		name           string
		api            *testhelper.TestApi
		body           []util.M
		expectedStatus int
		expectedBody   any
	}{
		{
			"update goals percentages",
			api,
			[]util.M{
				{"id": goals[0].ID, "percentage": 20},
				{"id": goals[1].ID, "percentage": 20},
				{"id": goals[2].ID, "percentage": 20},
				{"id": goals[3].ID, "percentage": 10},
				{"id": goals[4].ID, "percentage": 20},
				{"id": goals[5].ID, "percentage": 10},
			},
			200,
			[]util.M{
				formatGoal(goals[0], 20),
				formatGoal(goals[1], 20),
				formatGoal(goals[2], 20),
				formatGoal(goals[3], 10),
				formatGoal(goals[4], 20),
				formatGoal(goals[5], 10),
			},
		},
		{
			"missing goals",
			api,
			[]util.M{
				{"id": goals[0].ID, "percentage": 20},
				{"id": goals[1].ID, "percentage": 20},
				{"id": goals[2].ID, "percentage": 20},
				{"id": goals[3].ID, "percentage": 20},
				{"id": goals[4].ID, "percentage": 20},
			},
			400,
			util.M{"error": "one or more goals are missing"},
		},
		{
			"goals length matches but one is still missing",
			api,
			[]util.M{
				{"id": goals[0].ID, "percentage": 20},
				{"id": goals[1].ID, "percentage": 20},
				{"id": goals[2].ID, "percentage": 20},
				{"id": goals[3].ID, "percentage": 20},
				{"id": goals[4].ID, "percentage": 20},
				{"id": goals[4].ID, "percentage": 0},
			},
			400,
			util.M{"error": fmt.Sprintf("missing goal with id %d", goals[5].ID)},
		},
		{
			"ignore goals of another user",
			anotherUserApi,
			[]util.M{
				{"id": goals[6].ID, "percentage": 20},
				{"id": goals[7].ID, "percentage": 20},
				{"id": goals[8].ID, "percentage": 20},
				{"id": goals[9].ID, "percentage": 20},
				{"id": goals[10].ID, "percentage": 20},
				// first user goal
				{"id": goals[2].ID, "percentage": 0},
			},
			400,
			util.M{"error": fmt.Sprintf("missing goal with id %d", goals[11].ID)},
		},
		{
			"percentage lesser than 0",
			api,
			[]util.M{
				{"id": goals[0].ID, "percentage": -1},
				{"id": goals[1].ID, "percentage": 20},
				{"id": goals[2].ID, "percentage": 20},
				{"id": goals[3].ID, "percentage": 20},
				{"id": goals[4].ID, "percentage": 20},
				{"id": goals[5].ID, "percentage": 0},
			},
			400,
			util.M{"error": fmt.Sprintf("invalid percentage for goal id %d, it must be between 1 and 100", goals[0].ID)},
		},
		{
			"percentage greater than 100",
			api,
			[]util.M{
				{"id": goals[0].ID, "percentage": 101},
				{"id": goals[1].ID, "percentage": 20},
				{"id": goals[2].ID, "percentage": 20},
				{"id": goals[3].ID, "percentage": 20},
				{"id": goals[4].ID, "percentage": 20},
				{"id": goals[5].ID, "percentage": 0},
			},
			400,
			util.M{"error": fmt.Sprintf("invalid percentage for goal id %d, it must be between 1 and 100", goals[0].ID)},
		},
		{
			"sum of percentages greater than 100",
			api,
			[]util.M{
				{"id": goals[0].ID, "percentage": 20},
				{"id": goals[1].ID, "percentage": 20},
				{"id": goals[2].ID, "percentage": 20},
				{"id": goals[3].ID, "percentage": 20},
				{"id": goals[4].ID, "percentage": 20},
				{"id": goals[5].ID, "percentage": 1},
			},
			400,
			util.M{"error": "the sum of all percentages must be equal to 100"},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			var respBody any

			resp := d.api.Test(http.MethodPost, "/api/goals", d.body)
			d.api.UnmarshalBody(resp.Body, &respBody)

			assert.Equal(d.expectedStatus, resp.StatusCode)

			if reflect.TypeOf(d.expectedBody).Kind() == reflect.Slice {
				assert.ElementsMatch(d.expectedBody, respBody)
			} else {
				assert.Equal(d.expectedBody, respBody)
			}
		})
	}

	// assert goals of first user has been updated
	var respBody []util.M
	resp := api.Test(http.MethodGet, "/api/goals")
	api.UnmarshalBody(resp.Body, &respBody)
	assert.ElementsMatch([]util.M{
		formatGoal(goals[0], 20),
		formatGoal(goals[1], 20),
		formatGoal(goals[2], 20),
		formatGoal(goals[3], 10),
		formatGoal(goals[4], 20),
		formatGoal(goals[5], 10),
	}, respBody)

	// assert goals of second user has not been updated
	resp = anotherUserApi.Test(http.MethodGet, "/api/goals")
	api.UnmarshalBody(resp.Body, &respBody)
	assert.ElementsMatch([]util.M{
		formatGoal(goals[6], goals[6].Percentage),
		formatGoal(goals[7], goals[7].Percentage),
		formatGoal(goals[8], goals[8].Percentage),
		formatGoal(goals[9], goals[9].Percentage),
		formatGoal(goals[10], goals[10].Percentage),
		formatGoal(goals[11], goals[11].Percentage),
	}, respBody)
}
