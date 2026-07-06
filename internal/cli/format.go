package cli

import "fmt"

// fmtDuration renders seconds as MM:SS, or H:MM:SS once it reaches an hour.
func fmtDuration(seconds int64) string {
	if seconds < 0 {
		seconds = 0
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	sec := seconds % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, sec)
	}
	return fmt.Sprintf("%02d:%02d", m, sec)
}

// percentage returns elapsed/estimated*100 capped at 100, or 0 if no estimate.
func percentage(elapsedSeconds int64, estimatedMinutes int) int {
	if estimatedMinutes <= 0 {
		return 0
	}
	p := int((elapsedSeconds * 100) / int64(estimatedMinutes*60))
	if p > 100 {
		p = 100
	}
	if p < 0 {
		p = 0
	}
	return p
}
