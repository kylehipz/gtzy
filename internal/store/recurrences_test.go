package store

import (
	"path/filepath"
	"testing"
	"time"

	"gtzy/internal/db"
	"gtzy/internal/models"
)

func newTestDB(t *testing.T) *RecurrenceStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := db.Open(path)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return &RecurrenceStore{DB: conn}
}

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("parse date %q: %v", s, err)
	}
	return d
}

func TestRuleFiresDaily(t *testing.T) {
	rec := models.Recurrence{Freq: "daily", Interval: 2, StartDate: "2026-07-01"}

	cases := []struct {
		date string
		want bool
	}{
		{"2026-06-30", false}, // before start
		{"2026-07-01", true},  // day 0
		{"2026-07-02", false}, // day 1
		{"2026-07-03", true},  // day 2
		{"2026-07-05", true},  // day 4
	}
	for _, c := range cases {
		got, err := ruleFires(rec, mustDate(t, c.date))
		if err != nil {
			t.Fatalf("ruleFires(%s): %v", c.date, err)
		}
		if got != c.want {
			t.Errorf("daily interval=2 start=2026-07-01 on %s: got %v, want %v", c.date, got, c.want)
		}
	}
}

func TestRuleFiresWeekly(t *testing.T) {
	// start_date is a Wednesday (2026-07-01), rule fires Mon/Wed/Fri (1,3,5), every week.
	rec := models.Recurrence{Freq: "weekly", Interval: 1, DaysOfWeek: "1,3,5", StartDate: "2026-07-01"}

	cases := []struct {
		date string
		want bool
	}{
		{"2026-07-01", true},  // Wed, matches
		{"2026-07-02", false}, // Thu, not in list
		{"2026-07-03", true},  // Fri, matches
		{"2026-07-06", true},  // Mon, matches
		{"2026-07-04", false}, // Sat
	}
	for _, c := range cases {
		got, err := ruleFires(rec, mustDate(t, c.date))
		if err != nil {
			t.Fatalf("ruleFires(%s): %v", c.date, err)
		}
		if got != c.want {
			t.Errorf("weekly Mon/Wed/Fri on %s: got %v, want %v", c.date, got, c.want)
		}
	}
}

func TestRuleFiresWeeklyInterval(t *testing.T) {
	// Every 2nd week, on Mondays, starting the week of 2026-07-06 (a Monday).
	rec := models.Recurrence{Freq: "weekly", Interval: 2, DaysOfWeek: "1", StartDate: "2026-07-06"}

	cases := []struct {
		date string
		want bool
	}{
		{"2026-07-06", true},  // week 0
		{"2026-07-13", false}, // week 1 (skipped)
		{"2026-07-20", true},  // week 2
	}
	for _, c := range cases {
		got, err := ruleFires(rec, mustDate(t, c.date))
		if err != nil {
			t.Fatalf("ruleFires(%s): %v", c.date, err)
		}
		if got != c.want {
			t.Errorf("weekly interval=2 on %s: got %v, want %v", c.date, got, c.want)
		}
	}
}

func TestRuleFiresMonthlyWithClamping(t *testing.T) {
	day31 := 31
	rec := models.Recurrence{Freq: "monthly", Interval: 1, DayOfMonth: &day31, StartDate: "2026-01-31"}

	// February 2026 has 28 days, so day 31 clamps to the 28th.
	got, err := ruleFires(rec, mustDate(t, "2026-02-28"))
	if err != nil {
		t.Fatalf("ruleFires: %v", err)
	}
	if !got {
		t.Error("expected monthly rule clamped to Feb 28 to fire on 2026-02-28")
	}

	got, err = ruleFires(rec, mustDate(t, "2026-02-27"))
	if err != nil {
		t.Fatalf("ruleFires: %v", err)
	}
	if got {
		t.Error("expected monthly rule NOT to fire on 2026-02-27")
	}

	// March has 31 days, so it should fire on the 31st again.
	got, err = ruleFires(rec, mustDate(t, "2026-03-31"))
	if err != nil {
		t.Fatalf("ruleFires: %v", err)
	}
	if !got {
		t.Error("expected monthly rule to fire on 2026-03-31")
	}
}

func TestRuleFiresMonthlyInterval(t *testing.T) {
	day15 := 15
	rec := models.Recurrence{Freq: "monthly", Interval: 3, DayOfMonth: &day15, StartDate: "2026-01-15"}

	cases := []struct {
		date string
		want bool
	}{
		{"2026-01-15", true},  // month 0
		{"2026-02-15", false}, // month 1, skipped
		{"2026-03-15", false}, // month 2, skipped
		{"2026-04-15", true},  // month 3
	}
	for _, c := range cases {
		got, err := ruleFires(rec, mustDate(t, c.date))
		if err != nil {
			t.Fatalf("ruleFires(%s): %v", c.date, err)
		}
		if got != c.want {
			t.Errorf("monthly interval=3 on %s: got %v, want %v", c.date, got, c.want)
		}
	}
}

func TestEnsureInstancesForDateIdempotent(t *testing.T) {
	rs := newTestDB(t)

	_, err := rs.Create(RecurrenceInput{
		Title: "Daily habit", Freq: "daily", Interval: 1, StartDate: "2026-07-01",
	})
	if err != nil {
		t.Fatalf("create recurrence: %v", err)
	}

	n1, err := rs.EnsureInstancesForDate("2026-07-06")
	if err != nil {
		t.Fatalf("ensure 1st call: %v", err)
	}
	if n1 != 1 {
		t.Fatalf("expected 1 task created on first call, got %d", n1)
	}

	n2, err := rs.EnsureInstancesForDate("2026-07-06")
	if err != nil {
		t.Fatalf("ensure 2nd call: %v", err)
	}
	if n2 != 0 {
		t.Fatalf("expected 0 new tasks created on repeated call (idempotent), got %d", n2)
	}

	var count int
	if err := rs.DB.QueryRow(`SELECT COUNT(*) FROM tasks WHERE scheduled_date = ?`, "2026-07-06").Scan(&count); err != nil {
		t.Fatalf("count tasks: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 materialized task, got %d", count)
	}
}
