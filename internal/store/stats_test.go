package store

import (
	"path/filepath"
	"testing"

	"gtzy/internal/db"
)

func newStatsTestDBs(t *testing.T) (*TaskStore, *StatsStore) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := db.Open(path)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return &TaskStore{DB: conn}, &StatsStore{DB: conn, Recurrences: &RecurrenceStore{DB: conn}}
}

func TestStatsRangeAggregation(t *testing.T) {
	ts, ss := newStatsTestDBs(t)

	d1 := "2026-07-01"
	d2 := "2026-07-02"

	t1, err := ts.Create(TaskInput{Title: "A", ScheduledDate: &d1, EstimatedSeconds: 1800})
	if err != nil {
		t.Fatalf("create t1: %v", err)
	}
	t2, err := ts.Create(TaskInput{Title: "B", ScheduledDate: &d1, EstimatedSeconds: 1200})
	if err != nil {
		t.Fatalf("create t2: %v", err)
	}
	t3, err := ts.Create(TaskInput{Title: "C", ScheduledDate: &d2, EstimatedSeconds: 600})
	if err != nil {
		t.Fatalf("create t3: %v", err)
	}

	doneStatus := "done"
	if _, err := ts.Update(t1.ID, TaskPatch{Status: &doneStatus}); err != nil {
		t.Fatalf("complete t1: %v", err)
	}
	_ = t2
	_ = t3

	stats, err := ss.Range("2026-07-01", "2026-07-31", "")
	if err != nil {
		t.Fatalf("range: %v", err)
	}

	if stats.TasksTotal != 3 {
		t.Errorf("expected tasks_total=3, got %d", stats.TasksTotal)
	}
	if stats.TasksCompleted != 1 {
		t.Errorf("expected tasks_completed=1, got %d", stats.TasksCompleted)
	}
	if stats.EstimatedSecondsTotal != 3600 {
		t.Errorf("expected estimated_seconds_total=3600, got %d", stats.EstimatedSecondsTotal)
	}
	wantRate := 1.0 / 3.0
	if stats.CompletionRate < wantRate-0.001 || stats.CompletionRate > wantRate+0.001 {
		t.Errorf("expected completion_rate~%.3f, got %.3f", wantRate, stats.CompletionRate)
	}
	if len(stats.EstVsActualPerDay) != 2 {
		t.Errorf("expected 2 days in est_vs_actual, got %d", len(stats.EstVsActualPerDay))
	}
}

func TestStatsRangeEmpty(t *testing.T) {
	_, ss := newStatsTestDBs(t)

	stats, err := ss.Range("2026-01-01", "2026-01-31", "")
	if err != nil {
		t.Fatalf("range: %v", err)
	}
	if stats.TasksTotal != 0 || stats.CompletionRate != 0 {
		t.Errorf("expected zero-value stats for empty range, got %+v", stats)
	}
}
