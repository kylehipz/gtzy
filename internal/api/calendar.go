package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"gtzy/internal/store"
)

func (s *Server) registerCalendarRoutes(r chi.Router) {
	cs := &store.CalendarStore{DB: s.DB, Recurrences: &store.RecurrenceStore{DB: s.DB}}

	r.Get("/calendar/month", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		now := time.Now()
		year := now.Year()
		month := int(now.Month())
		if y := q.Get("year"); y != "" {
			if v, err := strconv.Atoi(y); err == nil {
				year = v
			}
		}
		if m := q.Get("month"); m != "" {
			if v, err := strconv.Atoi(m); err == nil {
				month = v
			}
		}
		days, err := cs.Month(year, month)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, days)
	})
}
