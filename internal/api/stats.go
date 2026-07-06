package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"gtzy/internal/store"
)

func (s *Server) registerStatsRoutes(r chi.Router) {
	ss := &store.StatsStore{DB: s.DB, Recurrences: &store.RecurrenceStore{DB: s.DB}}

	r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		from := q.Get("from")
		to := q.Get("to")
		categoryID := q.Get("category_id")
		if from == "" {
			from = "0000-01-01"
		}
		if to == "" {
			to = "9999-12-31"
		}
		stats, err := ss.Range(from, to, categoryID)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, stats)
	})
}
