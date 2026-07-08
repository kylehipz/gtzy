package store

import (
	"path/filepath"
	"testing"

	"gtzy/internal/db"
)

func newBloodSugarStore(t *testing.T) *BloodSugarStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := db.Open(path)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return &BloodSugarStore{DB: conn}
}

func seq(n int64) *int64 { return &n }

// CreateMany must be idempotent on meter records: re-syncing the same sequence
// numbers inserts nothing new, so callers can safely re-run sync.
func TestBloodSugarCreateMany_DedupBySeq(t *testing.T) {
	s := newBloodSugarStore(t)

	batch := []BloodSugarInput{
		{ValueMgdl: 100, TakenAt: "2026-07-08T08:00:00Z", Source: "meter", SeqNumber: seq(1)},
		{ValueMgdl: 142, TakenAt: "2026-07-08T12:00:00Z", Source: "meter", SeqNumber: seq(2)},
	}
	n, err := s.CreateMany(batch)
	if err != nil {
		t.Fatalf("CreateMany: %v", err)
	}
	if n != 2 {
		t.Fatalf("first sync inserted %d, want 2", n)
	}

	// Re-sync the same two plus one new record; only the new one should insert.
	again := append(batch, BloodSugarInput{ValueMgdl: 90, TakenAt: "2026-07-08T18:00:00Z", Source: "meter", SeqNumber: seq(3)})
	n, err = s.CreateMany(again)
	if err != nil {
		t.Fatalf("CreateMany (re-sync): %v", err)
	}
	if n != 1 {
		t.Fatalf("re-sync inserted %d, want 1 (seq 1 and 2 are duplicates)", n)
	}

	all, err := s.List("", "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("total readings = %d, want 3", len(all))
	}
}

// MaxMeterSeq drives incremental sync and must ignore manual readings.
func TestBloodSugarMaxMeterSeq(t *testing.T) {
	s := newBloodSugarStore(t)

	if got, err := s.MaxMeterSeq(); err != nil || got != 0 {
		t.Fatalf("MaxMeterSeq on empty = %d, %v; want 0, nil", got, err)
	}

	if _, err := s.Create(BloodSugarInput{ValueMgdl: 110, TakenAt: "2026-07-08T09:00:00Z", Source: "manual"}); err != nil {
		t.Fatalf("Create manual: %v", err)
	}
	if _, err := s.CreateMany([]BloodSugarInput{
		{ValueMgdl: 130, TakenAt: "2026-07-08T10:00:00Z", Source: "meter", SeqNumber: seq(7)},
		{ValueMgdl: 120, TakenAt: "2026-07-08T11:00:00Z", Source: "meter", SeqNumber: seq(4)},
	}); err != nil {
		t.Fatalf("CreateMany: %v", err)
	}

	got, err := s.MaxMeterSeq()
	if err != nil {
		t.Fatalf("MaxMeterSeq: %v", err)
	}
	if got != 7 {
		t.Errorf("MaxMeterSeq = %d, want 7 (max meter seq, manual ignored)", got)
	}
}
