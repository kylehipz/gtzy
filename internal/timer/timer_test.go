package timer

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"gtzy/internal/db"
	"gtzy/internal/store"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := db.Open(path)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func mustCreateTask(t *testing.T, ts *store.TaskStore, title, date, priority string) int64 {
	t.Helper()
	task, err := ts.Create(store.TaskInput{
		Title: title, Priority: priority, ScheduledDate: &date,
	})
	if err != nil {
		t.Fatalf("create task %q: %v", title, err)
	}
	return task.ID
}

func TestStartPauseAccumulation(t *testing.T) {
	conn := newTestDB(t)
	ts := &store.TaskStore{DB: conn}
	svc := New(conn)

	id := mustCreateTask(t, ts, "Focus work", store.TodayLocal(), "medium")

	if _, err := svc.Start(id); err != nil {
		t.Fatalf("start: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)

	task, err := svc.Pause(id)
	if err != nil {
		t.Fatalf("pause: %v", err)
	}
	if task.ActiveStartedAt != nil {
		t.Fatalf("expected active_started_at cleared after pause, got %v", *task.ActiveStartedAt)
	}
	if task.ActualSeconds < 1 {
		t.Fatalf("expected actual_seconds >= 1 after ~1.1s active, got %d", task.ActualSeconds)
	}
	if task.Status != "paused" {
		t.Fatalf("expected status=paused, got %q", task.Status)
	}

	// Start again and accumulate more.
	if _, err := svc.Start(id); err != nil {
		t.Fatalf("re-start: %v", err)
	}
	time.Sleep(600 * time.Millisecond)
	task, err = svc.Complete(id)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if task.Status != "done" {
		t.Fatalf("expected status=done, got %q", task.Status)
	}
	if task.ActualSeconds < 1 {
		t.Fatalf("expected accumulated actual_seconds >= 1 total, got %d", task.ActualSeconds)
	}
	if task.CompletedAt == nil {
		t.Fatal("expected completed_at to be set")
	}
}

func TestOneAtATimeInvariant(t *testing.T) {
	conn := newTestDB(t)
	ts := &store.TaskStore{DB: conn}
	svc := New(conn)

	date := store.TodayLocal()
	id1 := mustCreateTask(t, ts, "Task A", date, "medium")
	id2 := mustCreateTask(t, ts, "Task B", date, "medium")

	if _, err := svc.Start(id1); err != nil {
		t.Fatalf("start id1: %v", err)
	}
	if _, err := svc.Start(id2); err != nil {
		t.Fatalf("start id2: %v", err)
	}

	t1, err := ts.Get(id1)
	if err != nil {
		t.Fatalf("get id1: %v", err)
	}
	t2, err := ts.Get(id2)
	if err != nil {
		t.Fatalf("get id2: %v", err)
	}

	if t1.ActiveStartedAt != nil {
		t.Fatal("expected task 1 to be paused after starting task 2")
	}
	if t1.Status != "paused" {
		t.Fatalf("expected task 1 status=paused, got %q", t1.Status)
	}
	if t2.ActiveStartedAt == nil {
		t.Fatal("expected task 2 to be active")
	}

	var activeCount int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM tasks WHERE active_started_at IS NOT NULL`).Scan(&activeCount); err != nil {
		t.Fatalf("count active: %v", err)
	}
	if activeCount != 1 {
		t.Fatalf("expected exactly 1 active task, got %d", activeCount)
	}
}

func TestNextOrdering(t *testing.T) {
	conn := newTestDB(t)
	ts := &store.TaskStore{DB: conn}
	svc := New(conn)

	date := store.TodayLocal()
	// Created in an order that does NOT match priority, to prove ordering
	// comes from the priority/scheduled_start rules, not creation order.
	_ = mustCreateTask(t, ts, "Low prio", date, "low")
	urgentID := mustCreateTask(t, ts, "Urgent prio", date, "urgent")
	_ = mustCreateTask(t, ts, "Medium prio", date, "medium")

	cur, err := svc.Next()
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if cur == nil {
		t.Fatal("expected a current task after Next(), got nil")
	}
	if cur.ID != urgentID {
		t.Fatalf("expected urgent-priority task to be picked first, got %q", cur.Title)
	}

	// Advancing again should pause the urgent task and pick the next
	// eligible one (medium, since it was created before low but low is
	// lower priority... actually next eligible by priority is medium).
	cur2, err := svc.Next()
	if err != nil {
		t.Fatalf("next 2: %v", err)
	}
	if cur2 == nil {
		t.Fatal("expected a current task after second Next(), got nil")
	}
	if cur2.ID == urgentID {
		t.Fatal("expected Next() to advance away from the just-paused urgent task")
	}
	if cur2.Title != "Medium prio" {
		t.Fatalf("expected medium-priority task next, got %q", cur2.Title)
	}
}

func TestNextReturnsNilWhenNoneEligible(t *testing.T) {
	conn := newTestDB(t)
	svc := New(conn)

	cur, err := svc.Next()
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if cur != nil {
		t.Fatalf("expected nil current with no tasks, got %+v", cur)
	}
}
