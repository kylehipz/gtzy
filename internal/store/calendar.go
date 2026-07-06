package store

import (
	"database/sql"
	"fmt"
	"time"

	"gtzy/internal/models"
)

type CalendarStore struct {
	DB          *sql.DB
	Recurrences *RecurrenceStore
}

// Month returns per-day aggregates for every day in the given year/month,
// materializing recurrence instances for each visible day first.
func (s *CalendarStore) Month(year, month int, categoryID string) ([]models.CalendarDay, error) {
	first := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	daysInMonth := first.AddDate(0, 1, -1).Day()
	today := TodayLocal()

	days := make([]models.CalendarDay, 0, daysInMonth)
	for d := 1; d <= daysInMonth; d++ {
		date := time.Date(year, time.Month(month), d, 0, 0, 0, 0, time.Local).Format("2006-01-02")

		if _, err := s.Recurrences.EnsureInstancesForDate(date); err != nil {
			return nil, fmt.Errorf("materialize %s: %w", date, err)
		}

		var total, done int
		query := `SELECT COUNT(*), COALESCE(SUM(CASE WHEN status = 'done' THEN 1 ELSE 0 END), 0)
			 FROM tasks WHERE scheduled_date = ?`
		args := []any{date}
		if categoryID != "" {
			query += ` AND category_id = ?`
			args = append(args, categoryID)
		}
		if err := s.DB.QueryRow(query, args...).Scan(&total, &done); err != nil {
			return nil, err
		}

		state := "empty"
		ratio := 0.0
		if total > 0 {
			ratio = float64(done) / float64(total)
			switch {
			case done == total:
				state = "complete"
			case date <= today:
				state = "missed"
			default:
				state = "partial"
			}
		}

		days = append(days, models.CalendarDay{Date: date, Total: total, Done: done, Ratio: ratio, State: state})
	}
	return days, nil
}
