package testhelper

import (
	"time"

	"github.com/joaopsramos/fincon/internal/domain"
	"github.com/joaopsramos/fincon/internal/util"
)

func MiddleOfMonth() time.Time {
	now := time.Now().UTC()
	// Use middle of month to avoid test errors when subtracting/adding months
	return time.Date(now.Year(), now.Month(), 15, 0, 0, 0, 0, time.UTC)
}

func DateToJsonString(date time.Time) string {
	return date.Format("2006-01-02T15:04:05.999999Z")
}

func FormatExpense(e domain.Expense, g domain.Goal) util.M {
	return util.M{
		"id":      float64(e.ID),
		"name":    e.Name,
		"value":   float64(e.Value) / 100,
		"date":    DateToJsonString(e.Date),
		"goal_id": float64(g.ID),
	}
}
