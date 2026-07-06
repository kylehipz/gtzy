package timer

import (
	"database/sql"
	"errors"
	"time"

	"gtzy/internal/models"
	"gtzy/internal/store"
)

var ErrNotFound = errors.New("task not found")

// Service owns all timer state mutations, enforcing the one-active-task
// invariant. All timer mutations must go through here.
type Service struct {
	DB    *sql.DB
	Tasks *store.TaskStore
}

func New(db *sql.DB) *Service {
	return &Service{DB: db, Tasks: &store.TaskStore{DB: db}}
}

// Current returns the currently active task (with live elapsed_seconds), or
// nil if nothing is running.
func (s *Service) Current() (*models.Task, error) {
	var id int64
	err := s.DB.QueryRow(`SELECT id FROM tasks WHERE active_started_at IS NOT NULL LIMIT 1`).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t, err := s.Tasks.Get(id)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Start pauses whatever task is currently active (if different), then starts
// the target task and opens a new timer_sessions row for it.
func (s *Service) Start(id int64) (models.Task, error) {
	t, err := s.Tasks.Get(id)
	if err == sql.ErrNoRows {
		return models.Task{}, ErrNotFound
	}
	if err != nil {
		return models.Task{}, err
	}

	if t.ActiveStartedAt != nil {
		// already the active task
		return s.Tasks.Get(id)
	}

	if err := s.pauseCurrentIfAny(id); err != nil {
		return models.Task{}, err
	}

	now := store.NowUTC()
	if _, err := s.DB.Exec(
		`UPDATE tasks SET active_started_at = ?, status = 'in_progress', updated_at = ? WHERE id = ?`,
		now, now, id,
	); err != nil {
		return models.Task{}, err
	}
	if _, err := s.DB.Exec(
		`INSERT INTO timer_sessions (task_id, started_at, ended_at, duration_seconds) VALUES (?, ?, NULL, 0)`,
		id, now,
	); err != nil {
		return models.Task{}, err
	}
	return s.Tasks.Get(id)
}

// pauseCurrentIfAny pauses whichever task is currently active, if any, unless
// it's already the given id (caller handles that case separately).
func (s *Service) pauseCurrentIfAny(exceptID int64) error {
	var activeID int64
	err := s.DB.QueryRow(`SELECT id FROM tasks WHERE active_started_at IS NOT NULL LIMIT 1`).Scan(&activeID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if activeID == exceptID {
		return nil
	}
	return s.pauseTask(activeID, "paused")
}

// pauseTask computes elapsed delta, folds it into actual_seconds, closes the
// open session, clears active_started_at, and sets the given resulting status.
func (s *Service) pauseTask(id int64, resultStatus string) error {
	var activeStartedAt sql.NullString
	if err := s.DB.QueryRow(`SELECT active_started_at FROM tasks WHERE id = ?`, id).Scan(&activeStartedAt); err != nil {
		return err
	}
	if !activeStartedAt.Valid {
		// not active; just ensure status if requested
		if resultStatus != "" {
			_, err := s.DB.Exec(`UPDATE tasks SET status = ?, updated_at = ? WHERE id = ?`, resultStatus, store.NowUTC(), id)
			return err
		}
		return nil
	}

	started, err := store.ParseUTC(activeStartedAt.String)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	delta := int64(now.Sub(started).Seconds())
	if delta < 0 {
		delta = 0
	}
	nowStr := now.Format(time.RFC3339)

	if _, err := s.DB.Exec(
		`UPDATE tasks SET actual_seconds = actual_seconds + ?, active_started_at = NULL,
			status = ?, updated_at = ? WHERE id = ?`,
		delta, resultStatus, nowStr, id,
	); err != nil {
		return err
	}

	_, err = s.DB.Exec(
		`UPDATE timer_sessions SET ended_at = ?, duration_seconds = ?
		 WHERE task_id = ? AND ended_at IS NULL`,
		nowStr, delta, id,
	)
	return err
}

// Pause pauses the given task if it is currently active.
func (s *Service) Pause(id int64) (models.Task, error) {
	if _, err := s.Tasks.Get(id); err == sql.ErrNoRows {
		return models.Task{}, ErrNotFound
	} else if err != nil {
		return models.Task{}, err
	}
	if err := s.pauseTask(id, "paused"); err != nil {
		return models.Task{}, err
	}
	return s.Tasks.Get(id)
}

// PauseCurrent pauses whatever task is currently active, if any. No-op if
// nothing is running.
func (s *Service) PauseCurrent() (*models.Task, error) {
	cur, err := s.Current()
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, nil
	}
	if err := s.pauseTask(cur.ID, "paused"); err != nil {
		return nil, err
	}
	t, err := s.Tasks.Get(cur.ID)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Complete folds any active elapsed time into actual_seconds, then marks the
// task done.
func (s *Service) Complete(id int64) (models.Task, error) {
	if _, err := s.Tasks.Get(id); err == sql.ErrNoRows {
		return models.Task{}, ErrNotFound
	} else if err != nil {
		return models.Task{}, err
	}
	if err := s.pauseTask(id, "done"); err != nil {
		return models.Task{}, err
	}
	now := store.NowUTC()
	if _, err := s.DB.Exec(`UPDATE tasks SET completed_at = ?, updated_at = ? WHERE id = ?`, now, now, id); err != nil {
		return models.Task{}, err
	}
	return s.Tasks.Get(id)
}

// Next pauses the current active task (if any), then starts the next
// eligible task for today, ordered by scheduled_start ASC NULLS LAST, then
// priority (urgent>high>medium>low), then sort_order, then created_at.
// Returns the newly-current task, or nil if none eligible.
func (s *Service) Next() (*models.Task, error) {
	cur, err := s.Current()
	if err != nil {
		return nil, err
	}
	var justPausedID int64 = -1
	if cur != nil {
		if err := s.pauseTask(cur.ID, "paused"); err != nil {
			return nil, err
		}
		justPausedID = cur.ID
	}

	today := store.TodayLocal()
	row := s.DB.QueryRow(
		`SELECT id FROM tasks
		 WHERE scheduled_date = ? AND status != 'done' AND id != ?
		 ORDER BY
			scheduled_start IS NULL, scheduled_start ASC,
			CASE priority
				WHEN 'urgent' THEN 0
				WHEN 'high' THEN 1
				WHEN 'medium' THEN 2
				WHEN 'low' THEN 3
				ELSE 4
			END,
			sort_order ASC, created_at ASC
		 LIMIT 1`,
		today, justPausedID,
	)
	var nextID int64
	err = row.Scan(&nextID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	t, err := s.Start(nextID)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
