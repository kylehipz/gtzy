package store

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gtzy/internal/models"
)

type RecurrenceStore struct{ DB *sql.DB }

const recurCols = `id, title, notes, category_id, priority, estimated_minutes, scheduled_start,
	freq, interval, days_of_week, day_of_month, start_date, end_date, active, created_at, updated_at`

func scanRecurrence(row interface{ Scan(...any) error }) (models.Recurrence, error) {
	var rec models.Recurrence
	var active int
	err := row.Scan(
		&rec.ID, &rec.Title, &rec.Notes, &rec.CategoryID, &rec.Priority, &rec.EstimatedMinutes,
		&rec.ScheduledStart, &rec.Freq, &rec.Interval, &rec.DaysOfWeek, &rec.DayOfMonth,
		&rec.StartDate, &rec.EndDate, &active, &rec.CreatedAt, &rec.UpdatedAt,
	)
	rec.Active = active != 0
	return rec, err
}

func (s *RecurrenceStore) List() ([]models.Recurrence, error) {
	rows, err := s.DB.Query(`SELECT ` + recurCols + ` FROM recurrences ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recs := []models.Recurrence{}
	for rows.Next() {
		rec, err := scanRecurrence(rows)
		if err != nil {
			return nil, err
		}
		recs = append(recs, rec)
	}
	return recs, rows.Err()
}

func (s *RecurrenceStore) Get(id int64) (models.Recurrence, error) {
	row := s.DB.QueryRow(`SELECT `+recurCols+` FROM recurrences WHERE id = ?`, id)
	return scanRecurrence(row)
}

type RecurrenceInput struct {
	Title            string
	Notes            string
	CategoryID       *int64
	Priority         string
	EstimatedMinutes int
	ScheduledStart   *string
	Freq             string
	Interval         int
	DaysOfWeek       string
	DayOfMonth       *int
	StartDate        string
	EndDate          *string
}

func (s *RecurrenceStore) Create(in RecurrenceInput) (models.Recurrence, error) {
	if in.Priority == "" {
		in.Priority = "medium"
	}
	if in.Interval == 0 {
		in.Interval = 1
	}
	now := NowUTC()
	res, err := s.DB.Exec(
		`INSERT INTO recurrences (title, notes, category_id, priority, estimated_minutes,
			scheduled_start, freq, interval, days_of_week, day_of_month, start_date, end_date,
			active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		in.Title, in.Notes, in.CategoryID, in.Priority, in.EstimatedMinutes,
		in.ScheduledStart, in.Freq, in.Interval, in.DaysOfWeek, in.DayOfMonth,
		in.StartDate, in.EndDate, now, now,
	)
	if err != nil {
		return models.Recurrence{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return models.Recurrence{}, err
	}
	return s.Get(id)
}

type RecurrencePatch struct {
	Title            *string
	Notes            *string
	CategoryID       **int64
	Priority         *string
	EstimatedMinutes *int
	ScheduledStart   **string
	Freq             *string
	Interval         *int
	DaysOfWeek       *string
	DayOfMonth       **int
	StartDate        *string
	EndDate          **string
	Active           *bool
}

func (s *RecurrenceStore) Update(id int64, p RecurrencePatch) (models.Recurrence, error) {
	var sets []string
	var args []any
	add := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = ?", col))
		args = append(args, val)
	}

	if p.Title != nil {
		add("title", *p.Title)
	}
	if p.Notes != nil {
		add("notes", *p.Notes)
	}
	if p.CategoryID != nil {
		add("category_id", *p.CategoryID)
	}
	if p.Priority != nil {
		add("priority", *p.Priority)
	}
	if p.EstimatedMinutes != nil {
		add("estimated_minutes", *p.EstimatedMinutes)
	}
	if p.ScheduledStart != nil {
		add("scheduled_start", *p.ScheduledStart)
	}
	if p.Freq != nil {
		add("freq", *p.Freq)
	}
	if p.Interval != nil {
		add("interval", *p.Interval)
	}
	if p.DaysOfWeek != nil {
		add("days_of_week", *p.DaysOfWeek)
	}
	if p.DayOfMonth != nil {
		add("day_of_month", *p.DayOfMonth)
	}
	if p.StartDate != nil {
		add("start_date", *p.StartDate)
	}
	if p.EndDate != nil {
		add("end_date", *p.EndDate)
	}
	if p.Active != nil {
		v := 0
		if *p.Active {
			v = 1
		}
		add("active", v)
	}

	if len(sets) == 0 {
		return s.Get(id)
	}
	add("updated_at", NowUTC())
	args = append(args, id)
	q := `UPDATE recurrences SET ` + strings.Join(sets, ", ") + ` WHERE id = ?`
	if _, err := s.DB.Exec(q, args...); err != nil {
		return models.Recurrence{}, err
	}
	return s.Get(id)
}

// Delete sets active=0 (soft delete, keeps past instances via ON DELETE SET NULL).
// If hard is true, also removes future undone instances of this rule.
func (s *RecurrenceStore) Delete(id int64, hard bool) error {
	if _, err := s.DB.Exec(`UPDATE recurrences SET active = 0, updated_at = ? WHERE id = ?`, NowUTC(), id); err != nil {
		return err
	}
	if hard {
		today := TodayLocal()
		_, err := s.DB.Exec(
			`DELETE FROM tasks WHERE recurrence_id = ? AND status != 'done' AND scheduled_date >= ?`,
			id, today,
		)
		return err
	}
	return nil
}

// EnsureInstancesForDate materializes concrete task rows for every active
// recurrence whose rule fires on date (YYYY-MM-DD). Idempotent: relies on the
// unique index on (recurrence_id, scheduled_date) to avoid duplicates.
func (s *RecurrenceStore) EnsureInstancesForDate(date string) (int, error) {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return 0, fmt.Errorf("invalid date %q: %w", date, err)
	}

	rows, err := s.DB.Query(
		`SELECT ` + recurCols + ` FROM recurrences
		 WHERE active = 1 AND start_date <= ? AND (end_date IS NULL OR end_date >= ?)`,
		date, date,
	)
	if err != nil {
		return 0, err
	}
	var recs []models.Recurrence
	for rows.Next() {
		rec, err := scanRecurrence(rows)
		if err != nil {
			rows.Close()
			return 0, err
		}
		recs = append(recs, rec)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}

	created := 0
	now := NowUTC()
	for _, rec := range recs {
		fires, err := ruleFires(rec, d)
		if err != nil {
			return created, err
		}
		if !fires {
			continue
		}
		res, err := s.DB.Exec(
			`INSERT INTO tasks (title, notes, category_id, priority, status, estimated_minutes,
				scheduled_date, scheduled_start, recurrence_id, sort_order, created_at, updated_at)
			 VALUES (?, ?, ?, ?, 'todo', ?, ?, ?, ?, 0, ?, ?)
			 ON CONFLICT (recurrence_id, scheduled_date) WHERE recurrence_id IS NOT NULL DO NOTHING`,
			rec.Title, rec.Notes, rec.CategoryID, rec.Priority, rec.EstimatedMinutes,
			date, rec.ScheduledStart, rec.ID, now, now,
		)
		if err != nil {
			return created, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			created++
		}
	}
	return created, nil
}

func ruleFires(rec models.Recurrence, d time.Time) (bool, error) {
	start, err := time.Parse("2006-01-02", rec.StartDate)
	if err != nil {
		return false, fmt.Errorf("invalid start_date on recurrence %d: %w", rec.ID, err)
	}
	if d.Before(start) {
		return false, nil
	}
	interval := rec.Interval
	if interval <= 0 {
		interval = 1
	}

	switch rec.Freq {
	case "daily":
		days := int(d.Sub(start).Hours() / 24)
		return days%interval == 0, nil

	case "weekly":
		if !weekdayInList(int(d.Weekday()), rec.DaysOfWeek) {
			return false, nil
		}
		sundayOfStart := start.AddDate(0, 0, -int(start.Weekday()))
		sundayOfDate := d.AddDate(0, 0, -int(d.Weekday()))
		weeks := int(sundayOfDate.Sub(sundayOfStart).Hours() / 24 / 7)
		return weeks%interval == 0, nil

	case "monthly":
		monthOffset := (d.Year()*12 + int(d.Month())) - (start.Year()*12 + int(start.Month()))
		if monthOffset < 0 || monthOffset%interval != 0 {
			return false, nil
		}
		dayOfMonth := 1
		if rec.DayOfMonth != nil {
			dayOfMonth = *rec.DayOfMonth
		}
		lastDay := daysInMonth(d.Year(), int(d.Month()))
		if dayOfMonth > lastDay {
			dayOfMonth = lastDay
		}
		return d.Day() == dayOfMonth, nil

	default:
		return false, fmt.Errorf("unknown freq %q", rec.Freq)
	}
}

func weekdayInList(weekday int, csv string) bool {
	if csv == "" {
		return false
	}
	for _, part := range strings.Split(csv, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err == nil && n == weekday {
			return true
		}
	}
	return false
}

func daysInMonth(year, month int) int {
	return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
