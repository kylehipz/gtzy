package meter

import (
	"context"
	"database/sql"
	"log"
	"time"

	"gtzy/internal/models"
	"gtzy/internal/store"
)

// Watcher polls for the meter in the background and auto-imports new readings the
// moment the meter advertises (which it does when you press save). It reproduces
// the phone-app experience: check your blood sugar and the reading shows up on
// its own. Enable it from `gtzy serve` via the GTZY_METER_WATCH env var.
type Watcher struct {
	DB       *sql.DB
	Interval time.Duration                             // cooldown between scan cycles (default 10s)
	OnImport func(readings []models.BloodSugarReading) // called after each auto-import; nil-safe
}

// Start launches the watch loop in a goroutine that runs until ctx is cancelled
// (or the process exits).
func (w *Watcher) Start(ctx context.Context) {
	interval := w.Interval
	if interval <= 0 {
		interval = 10 * time.Second
	}
	go w.loop(ctx, interval)
}

func (w *Watcher) loop(ctx context.Context, interval time.Duration) {
	bs := &store.BloodSugarStore{DB: w.DB}
	var lastErr string // suppress repeated identical errors (e.g. adapter off)

	for {
		if ctx.Err() != nil {
			return
		}

		fetched, inserted, err := SyncInto(ctx, w.DB)
		switch {
		case err != nil:
			if msg := err.Error(); msg != lastErr {
				log.Printf("meter watch: %v", err)
				lastErr = msg
			}
		case inserted > 0:
			lastErr = ""
			log.Printf("meter watch: imported %d new reading(s) (%d fetched)", inserted, fetched)
			if w.OnImport != nil {
				if rows, lerr := bs.RecentMeter(inserted); lerr == nil {
					w.OnImport(rows)
				}
			}
		default:
			lastErr = ""
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}
