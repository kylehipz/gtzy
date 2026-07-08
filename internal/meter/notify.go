package meter

import (
	"fmt"
	"os/exec"
	"time"

	"gtzy/internal/models"
)

// NotifyImport is a Watcher.OnImport callback that fires a desktop notification
// summarizing what was just auto-synced, mirroring the meter phone app's
// "reading received" popup. It is best-effort: if notify-send is not on PATH
// (e.g. a headless host) it silently does nothing.
func NotifyImport(readings []models.BloodSugarReading) {
	if len(readings) == 0 {
		return
	}
	var body string
	if len(readings) == 1 {
		r := readings[0]
		when := r.TakenAt
		if t, err := time.Parse(time.RFC3339, r.TakenAt); err == nil {
			when = t.Local().Format("15:04")
		}
		body = fmt.Sprintf("%.0f mg/dL at %s", r.ValueMgdl, when)
	} else {
		body = fmt.Sprintf("imported %d readings", len(readings))
	}
	notifySend("gtzy: blood sugar synced", body)
}

func notifySend(title, body string) {
	path, err := exec.LookPath("notify-send")
	if err != nil {
		return
	}
	_ = exec.Command(path, title, body).Run()
}
