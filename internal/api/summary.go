package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"gtzy/internal/store"
)

func (s *Server) registerSummaryRoutes(r chi.Router) {
	sums := &store.SummaryStore{DB: s.DB}
	stats := &store.StatsStore{DB: s.DB, Recurrences: &store.RecurrenceStore{DB: s.DB}}
	tasks := &store.TaskStore{DB: s.DB}
	journal := &store.JournalStore{DB: s.DB}
	bloodSugar := &store.BloodSugarStore{DB: s.DB}

	r.Get("/summary", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		periodType, periodKey := q.Get("period"), q.Get("key")
		if periodType == "" || periodKey == "" {
			writeErr(w, http.StatusBadRequest, "period and key are required")
			return
		}

		if s.AI == nil || !s.AI.Enabled() {
			writeJSON(w, http.StatusOK, map[string]any{"enabled": false})
			return
		}

		sum, found, err := sums.Get(periodType, periodKey)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !found {
			writeJSON(w, http.StatusOK, map[string]any{"enabled": true, "cached": false})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"enabled": true, "cached": true, "summary": sum})
	})

	r.Post("/summary/generate", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		periodType, periodKey := q.Get("period"), q.Get("key")
		if periodType == "" || periodKey == "" {
			writeErr(w, http.StatusBadRequest, "period and key are required")
			return
		}
		if s.AI == nil || !s.AI.Enabled() {
			writeErr(w, http.StatusBadRequest, "AI summaries are disabled: no API key configured")
			return
		}

		from, to, err := periodRange(periodType, periodKey)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}

		periodStats, err := stats.Range(from, to, "")
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		periodTasks, err := tasks.List(store.TaskFilters{From: from, To: to})
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		entries, err := journal.List("", from, to)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}

		prompt := buildSummaryPrompt(periodType, periodKey, periodStats, periodTasks, entries)

		content, err := s.AI.Generate(periodType, periodKey, prompt)
		if err != nil {
			writeErr(w, http.StatusBadGateway, err.Error())
			return
		}

		sum, err := sums.Upsert(periodType, periodKey, content, s.AI.Model())
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"enabled": true, "cached": false, "summary": sum})
	})

	r.Post("/summary/journal", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		from, to := q.Get("from"), q.Get("to")
		if _, err := time.Parse("2006-01-02", from); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid from date, want YYYY-MM-DD")
			return
		}
		if _, err := time.Parse("2006-01-02", to); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid to date, want YYYY-MM-DD")
			return
		}
		if s.AI == nil || !s.AI.Enabled() {
			writeErr(w, http.StatusBadRequest, "AI summaries are disabled: no API key configured")
			return
		}

		entries, err := journal.List("", from, to)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}

		prompt := buildJournalSummaryPrompt(from, to, entries)

		key := from + ".." + to
		content, err := s.AI.Generate("journal", key, prompt)
		if err != nil {
			writeErr(w, http.StatusBadGateway, err.Error())
			return
		}

		sum, err := sums.Upsert("journal", key, content, s.AI.Model())
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"enabled": true, "cached": false, "summary": sum})
	})

	r.Post("/summary/bloodsugar", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		from, to := q.Get("from"), q.Get("to")
		if _, err := time.Parse("2006-01-02", from); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid from date, want YYYY-MM-DD")
			return
		}
		if _, err := time.Parse("2006-01-02", to); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid to date, want YYYY-MM-DD")
			return
		}
		if s.AI == nil || !s.AI.Enabled() {
			writeErr(w, http.StatusBadRequest, "AI summaries are disabled: no API key configured")
			return
		}

		// Include the whole "to" day by extending the upper bound past midnight.
		readings, err := bloodSugar.List(from, to+"T23:59:59Z")
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		stats := computeBloodSugarStats(readings)
		prompt := buildBloodSugarSummaryPrompt(from, to, readings, stats)

		key := from + ".." + to
		content, err := s.AI.Generate("blood_sugar", key, prompt)
		if err != nil {
			writeErr(w, http.StatusBadGateway, err.Error())
			return
		}

		sum, err := sums.Upsert("blood_sugar", key, content, s.AI.Model())
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"enabled": true, "cached": false, "summary": sum})
	})
}

func periodRange(periodType, periodKey string) (from, to string, err error) {
	switch periodType {
	case "day":
		if _, err := time.Parse("2006-01-02", periodKey); err != nil {
			return "", "", fmt.Errorf("invalid day key %q, want YYYY-MM-DD", periodKey)
		}
		return periodKey, periodKey, nil

	case "week":
		parts := strings.SplitN(periodKey, "-W", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid week key %q, want YYYY-Www", periodKey)
		}
		year, err1 := strconv.Atoi(parts[0])
		week, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			return "", "", fmt.Errorf("invalid week key %q, want YYYY-Www", periodKey)
		}
		jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC)
		jan4Weekday := int(jan4.Weekday())
		if jan4Weekday == 0 {
			jan4Weekday = 7
		}
		week1Monday := jan4.AddDate(0, 0, -(jan4Weekday - 1))
		monday := week1Monday.AddDate(0, 0, (week-1)*7)
		sunday := monday.AddDate(0, 0, 6)
		return monday.Format("2006-01-02"), sunday.Format("2006-01-02"), nil

	case "month":
		t, err := time.Parse("2006-01", periodKey)
		if err != nil {
			return "", "", fmt.Errorf("invalid month key %q, want YYYY-MM", periodKey)
		}
		firstDay := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		lastDay := firstDay.AddDate(0, 1, -1)
		return firstDay.Format("2006-01-02"), lastDay.Format("2006-01-02"), nil

	default:
		return "", "", fmt.Errorf("unknown period %q, want day, week, or month", periodType)
	}
}
