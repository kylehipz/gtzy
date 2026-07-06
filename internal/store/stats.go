package store

import (
	"database/sql"
	"sort"
	"time"

	"gtzy/internal/models"
)

type StatsStore struct {
	DB          *sql.DB
	Recurrences *RecurrenceStore
}

// Range computes aggregate analytics for tasks scheduled in [from, to]
// (inclusive, YYYY-MM-DD). An empty categoryID applies no category filter.
func (s *StatsStore) Range(from, to, categoryID string) (models.Stats, error) {
	var stats models.Stats

	q := `SELECT scheduled_date, status, estimated_minutes, actual_seconds
		 FROM tasks WHERE scheduled_date >= ? AND scheduled_date <= ?`
	args := []any{from, to}
	if categoryID != "" {
		q += ` AND category_id = ?`
		args = append(args, categoryID)
	}
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return stats, err
	}
	type dayAgg struct {
		total, done   int
		estMinutes    int
		actualSeconds int64
	}
	perDay := map[string]*dayAgg{}
	for rows.Next() {
		var date, status string
		var estMin int
		var actualSec int64
		if err := rows.Scan(&date, &status, &estMin, &actualSec); err != nil {
			rows.Close()
			return stats, err
		}
		agg, ok := perDay[date]
		if !ok {
			agg = &dayAgg{}
			perDay[date] = agg
		}
		agg.total++
		if status == "done" {
			agg.done++
		}
		agg.estMinutes += estMin
		agg.actualSeconds += actualSec

		stats.TasksTotal++
		if status == "done" {
			stats.TasksCompleted++
		}
		stats.EstimatedMinutesTotal += estMin
		stats.ActualSecondsTotal += actualSec
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return stats, err
	}

	if stats.TasksTotal > 0 {
		stats.CompletionRate = float64(stats.TasksCompleted) / float64(stats.TasksTotal)
	}

	dates := make([]string, 0, len(perDay))
	for d := range perDay {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	var completionSum float64
	var completionDays int
	for _, d := range dates {
		agg := perDay[d]
		stats.EstVsActualPerDay = append(stats.EstVsActualPerDay, models.EstVsActualDay{
			Date: d, EstimatedMinutes: agg.estMinutes, ActualSeconds: agg.actualSeconds,
			Total: agg.total, Done: agg.done,
		})
		if agg.total > 0 {
			completionSum += float64(agg.done) / float64(agg.total)
			completionDays++
		}
	}
	if completionDays > 0 {
		stats.AvgDailyCompletion = completionSum / float64(completionDays)
	}

	catQ := `SELECT t.category_id, COALESCE(c.name, 'Uncategorized'), COALESCE(SUM(t.actual_seconds), 0)
		 FROM tasks t LEFT JOIN categories c ON c.id = t.category_id
		 WHERE t.scheduled_date >= ? AND t.scheduled_date <= ?`
	catArgs := []any{from, to}
	if categoryID != "" {
		catQ += ` AND t.category_id = ?`
		catArgs = append(catArgs, categoryID)
	}
	catQ += ` GROUP BY t.category_id, c.name ORDER BY SUM(t.actual_seconds) DESC`
	catRows, err := s.DB.Query(catQ, catArgs...)
	if err != nil {
		return stats, err
	}
	defer catRows.Close()
	for catRows.Next() {
		var ct models.CategoryTime
		if err := catRows.Scan(&ct.CategoryID, &ct.CategoryName, &ct.Seconds); err != nil {
			return stats, err
		}
		stats.TimeByCategory = append(stats.TimeByCategory, ct)
	}
	if err := catRows.Err(); err != nil {
		return stats, err
	}
	if len(stats.TimeByCategory) > 0 {
		stats.BusiestCategory = stats.TimeByCategory[0].CategoryName
	}

	streak, err := s.currentStreak()
	if err != nil {
		return stats, err
	}
	stats.CurrentStreak = streak

	return stats, nil
}

// currentStreak counts consecutive complete days walking backward from today.
// If today isn't complete yet (in progress or empty), it's skipped so an
// unfinished today doesn't zero out a real streak from prior days.
func (s *StatsStore) currentStreak() (int, error) {
	dayState := func(date string) (string, error) {
		var total, done int
		if err := s.DB.QueryRow(
			`SELECT COUNT(*), COALESCE(SUM(CASE WHEN status = 'done' THEN 1 ELSE 0 END), 0)
			 FROM tasks WHERE scheduled_date = ?`, date,
		).Scan(&total, &done); err != nil {
			return "", err
		}
		if total == 0 {
			return "empty", nil
		}
		if done == total {
			return "complete", nil
		}
		return "incomplete", nil
	}

	loc := time.Now().Location()
	cursor := time.Now().In(loc)

	todayState, err := dayState(cursor.Format("2006-01-02"))
	if err != nil {
		return 0, err
	}
	if todayState != "complete" {
		cursor = cursor.AddDate(0, 0, -1)
	}

	streak := 0
	for {
		date := cursor.Format("2006-01-02")
		state, err := dayState(date)
		if err != nil {
			return streak, err
		}
		if state != "complete" {
			break
		}
		streak++
		cursor = cursor.AddDate(0, 0, -1)
	}
	return streak, nil
}
