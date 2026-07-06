package store

import "time"

// NowUTC returns the current instant as an ISO-8601 UTC string.
func NowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// TodayLocal returns today's date as YYYY-MM-DD in the server's local timezone.
func TodayLocal() string {
	return time.Now().Format("2006-01-02")
}

// ParseUTC parses an ISO-8601 UTC timestamp string as produced by NowUTC.
func ParseUTC(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
