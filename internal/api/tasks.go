package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"gtzy/internal/store"
)

func (s *Server) registerTaskRoutes(r chi.Router) {
	ts := &store.TaskStore{DB: s.DB}
	rs := &store.RecurrenceStore{DB: s.DB}

	r.Get("/tasks", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		f := store.TaskFilters{
			Date:       q.Get("date"),
			From:       q.Get("from"),
			To:         q.Get("to"),
			Status:     q.Get("status"),
			Priority:   q.Get("priority"),
			CategoryID: q.Get("category_id"),
		}
		if f.Date != "" {
			if _, err := rs.EnsureInstancesForDate(f.Date); err != nil {
				writeErr(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		tasks, err := ts.List(f)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, tasks)
	})

	r.Post("/tasks", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Title            string  `json:"title"`
			Notes            string  `json:"notes"`
			CategoryID       *int64  `json:"category_id"`
			Priority         string  `json:"priority"`
			EstimatedMinutes int     `json:"estimated_minutes"`
			ScheduledDate    *string `json:"scheduled_date"`
			ScheduledStart   *string `json:"scheduled_start"`
			SortOrder        int     `json:"sort_order"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid body")
			return
		}
		if body.Title == "" {
			writeErr(w, http.StatusBadRequest, "title is required")
			return
		}
		t, err := ts.Create(store.TaskInput{
			Title: body.Title, Notes: body.Notes, CategoryID: body.CategoryID,
			Priority: body.Priority, EstimatedMinutes: body.EstimatedMinutes,
			ScheduledDate: body.ScheduledDate, ScheduledStart: body.ScheduledStart,
			SortOrder: body.SortOrder,
		})
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, t)
	})

	r.Get("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid id")
			return
		}
		t, err := ts.Get(id)
		if err != nil {
			writeErr(w, http.StatusNotFound, "task not found")
			return
		}
		writeJSON(w, http.StatusOK, t)
	})

	r.Patch("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid id")
			return
		}
		var body struct {
			Title            *string `json:"title"`
			Notes            *string `json:"notes"`
			CategoryID       *int64  `json:"category_id"`
			Priority         *string `json:"priority"`
			Status           *string `json:"status"`
			EstimatedMinutes *int    `json:"estimated_minutes"`
			ScheduledDate    *string `json:"scheduled_date"`
			ScheduledStart   *string `json:"scheduled_start"`
			SortOrder        *int    `json:"sort_order"`
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

		patch := store.TaskPatch{
			Title: body.Title, Notes: body.Notes, Priority: body.Priority,
			Status: body.Status, EstimatedMinutes: body.EstimatedMinutes,
			SortOrder: body.SortOrder,
		}
		if hasKey(raw, "category_id") {
			patch.CategoryID = &body.CategoryID
		}
		if hasKey(raw, "scheduled_date") {
			patch.ScheduledDate = &body.ScheduledDate
		}
		if hasKey(raw, "scheduled_start") {
			patch.ScheduledStart = &body.ScheduledStart
		}

		t, err := ts.Update(id, patch)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, t)
	})

	r.Delete("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := ts.Delete(id); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	s.registerTimerActionRoutes(r, ts)
}
