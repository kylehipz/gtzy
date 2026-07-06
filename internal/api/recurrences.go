package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"gtzy/internal/store"
)

func (s *Server) registerRecurrenceRoutes(r chi.Router) {
	rs := &store.RecurrenceStore{DB: s.DB}

	r.Get("/recurrences", func(w http.ResponseWriter, r *http.Request) {
		recs, err := rs.List()
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, recs)
	})

	r.Post("/recurrences", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Title            string  `json:"title"`
			Notes            string  `json:"notes"`
			CategoryID       *int64  `json:"category_id"`
			Priority         string  `json:"priority"`
			EstimatedMinutes int     `json:"estimated_minutes"`
			ScheduledStart   *string `json:"scheduled_start"`
			Freq             string  `json:"freq"`
			Interval         int     `json:"interval"`
			DaysOfWeek       string  `json:"days_of_week"`
			DayOfMonth       *int    `json:"day_of_month"`
			StartDate        string  `json:"start_date"`
			EndDate          *string `json:"end_date"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid body")
			return
		}
		if body.Title == "" || body.Freq == "" || body.StartDate == "" {
			writeErr(w, http.StatusBadRequest, "title, freq, and start_date are required")
			return
		}
		rec, err := rs.Create(store.RecurrenceInput{
			Title: body.Title, Notes: body.Notes, CategoryID: body.CategoryID,
			Priority: body.Priority, EstimatedMinutes: body.EstimatedMinutes,
			ScheduledStart: body.ScheduledStart, Freq: body.Freq, Interval: body.Interval,
			DaysOfWeek: body.DaysOfWeek, DayOfMonth: body.DayOfMonth,
			StartDate: body.StartDate, EndDate: body.EndDate,
		})
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, rec)
	})

	r.Patch("/recurrences/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid id")
			return
		}
		var body struct {
			Title            *string `json:"title"`
			Notes            *string `json:"notes"`
			Priority         *string `json:"priority"`
			EstimatedMinutes *int    `json:"estimated_minutes"`
			Freq             *string `json:"freq"`
			Interval         *int    `json:"interval"`
			DaysOfWeek       *string `json:"days_of_week"`
			StartDate        *string `json:"start_date"`
			Active           *bool   `json:"active"`
		}
		raw, err := readAll(r)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid body")
			return
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid body")
			return
		}
		rec, err := rs.Update(id, store.RecurrencePatch{
			Title: body.Title, Notes: body.Notes, Priority: body.Priority,
			EstimatedMinutes: body.EstimatedMinutes, Freq: body.Freq, Interval: body.Interval,
			DaysOfWeek: body.DaysOfWeek, StartDate: body.StartDate, Active: body.Active,
		})
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, rec)
	})

	r.Delete("/recurrences/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid id")
			return
		}
		hard := r.URL.Query().Get("hard") == "1"
		if err := rs.Delete(id, hard); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
}
