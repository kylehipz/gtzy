package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gtzy/internal/models"
)

type TaskStore struct{ DB *sql.DB }

type TaskFilters struct {
	Date       string // exact scheduled_date match
	From       string // scheduled_date >= From
	To         string // scheduled_date <= To
	Status     string
	Priority   string
	CategoryID string // numeric string, empty = no filter
}

const taskCols = `id, title, notes, category_id, priority, status, estimated_seconds,
	actual_seconds, scheduled_date, scheduled_start, active_started_at, completed_at,
	recurrence_id, sort_order, created_at, updated_at`

func scanTask(row interface{ Scan(...any) error }) (models.Task, error) {
	var t models.Task
	err := row.Scan(
		&t.ID, &t.Title, &t.Notes, &t.CategoryID, &t.Priority, &t.Status, &t.EstimatedSeconds,
		&t.ActualSeconds, &t.ScheduledDate, &t.ScheduledStart, &t.ActiveStartedAt, &t.CompletedAt,
		&t.RecurrenceID, &t.SortOrder, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return t, err
	}
	computeElapsed(&t)
	return t, nil
}

func computeElapsed(t *models.Task) {
	t.ElapsedSeconds = t.ActualSeconds
	if t.ActiveStartedAt != nil {
		t.IsActive = true
		if started, err := ParseUTC(*t.ActiveStartedAt); err == nil {
			t.ElapsedSeconds += int64(time.Since(started).Seconds())
		}
	}
}

func (s *TaskStore) List(f TaskFilters) ([]models.Task, error) {
	q := `SELECT ` + taskCols + ` FROM tasks WHERE 1=1`
	var args []any

	if f.Date != "" {
		q += ` AND scheduled_date = ?`
		args = append(args, f.Date)
	}
	if f.From != "" {
		q += ` AND scheduled_date >= ?`
		args = append(args, f.From)
	}
	if f.To != "" {
		q += ` AND scheduled_date <= ?`
		args = append(args, f.To)
	}
	if f.Status != "" {
		q += ` AND status = ?`
		args = append(args, f.Status)
	}
	if f.Priority != "" {
		q += ` AND priority = ?`
		args = append(args, f.Priority)
	}
	if f.CategoryID != "" {
		q += ` AND category_id = ?`
		args = append(args, f.CategoryID)
	}
	q += ` ORDER BY scheduled_date IS NULL, scheduled_date, scheduled_start IS NULL, scheduled_start, sort_order, created_at`

	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.Task{}
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (s *TaskStore) Get(id int64) (models.Task, error) {
	row := s.DB.QueryRow(`SELECT `+taskCols+` FROM tasks WHERE id = ?`, id)
	return scanTask(row)
}

type TaskInput struct {
	Title            string
	Notes            string
	CategoryID       *int64
	Priority         string
	EstimatedSeconds int
	ScheduledDate    *string
	ScheduledStart   *string
	SortOrder        int
	RecurrenceID     *int64
}

func (s *TaskStore) Create(in TaskInput) (models.Task, error) {
	if in.Priority == "" {
		in.Priority = "medium"
	}
	now := NowUTC()
	res, err := s.DB.Exec(
		`INSERT INTO tasks (title, notes, category_id, priority, status, estimated_seconds,
			scheduled_date, scheduled_start, recurrence_id, sort_order, created_at, updated_at)
		 VALUES (?, ?, ?, ?, 'todo', ?, ?, ?, ?, ?, ?, ?)`,
		in.Title, in.Notes, in.CategoryID, in.Priority, in.EstimatedSeconds,
		in.ScheduledDate, in.ScheduledStart, in.RecurrenceID, in.SortOrder, now, now,
	)
	if err != nil {
		return models.Task{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return models.Task{}, err
	}
	return s.Get(id)
}

// TaskPatch holds optional fields for a partial update; nil means "don't touch".
type TaskPatch struct {
	Title            *string
	Notes            *string
	CategoryID       **int64
	Priority         *string
	Status           *string
	EstimatedSeconds *int
	ScheduledDate    **string
	ScheduledStart   **string
	SortOrder        *int
}

func (s *TaskStore) Update(id int64, p TaskPatch) (models.Task, error) {
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
	if p.Status != nil {
		add("status", *p.Status)
	}
	if p.EstimatedSeconds != nil {
		add("estimated_seconds", *p.EstimatedSeconds)
	}
	if p.ScheduledDate != nil {
		add("scheduled_date", *p.ScheduledDate)
	}
	if p.ScheduledStart != nil {
		add("scheduled_start", *p.ScheduledStart)
	}
	if p.SortOrder != nil {
		add("sort_order", *p.SortOrder)
	}

	if len(sets) == 0 {
		return s.Get(id)
	}

	add("updated_at", NowUTC())
	args = append(args, id)
	q := `UPDATE tasks SET ` + strings.Join(sets, ", ") + ` WHERE id = ?`
	if _, err := s.DB.Exec(q, args...); err != nil {
		return models.Task{}, err
	}
	return s.Get(id)
}

func (s *TaskStore) Delete(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	return err
}
