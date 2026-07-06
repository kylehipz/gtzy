package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"gtzy/internal/store"
)

func (s *Server) registerJournalRoutes(r chi.Router) {
	js := &store.JournalStore{DB: s.DB}

	r.Get("/journal", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		entries, err := js.List(q.Get("date"), q.Get("from"), q.Get("to"))
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, entries)
	})

	r.Post("/journal", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Date    string  `json:"date"`
			Title   string  `json:"title"`
			Content string  `json:"content"`
			Mood    *string `json:"mood"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid body")
			return
		}
		if body.Date == "" {
			writeErr(w, http.StatusBadRequest, "date is required")
			return
		}
		e, err := js.Create(store.JournalInput{Date: body.Date, Title: body.Title, Content: body.Content, Mood: body.Mood})
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, e)
	})

	r.Get("/journal/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid id")
			return
		}
		e, err := js.Get(id)
		if err != nil {
			writeErr(w, http.StatusNotFound, "journal entry not found")
			return
		}
		writeJSON(w, http.StatusOK, e)
	})

	r.Patch("/journal/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid id")
			return
		}
		var body struct {
			Title   *string `json:"title"`
			Content *string `json:"content"`
			Mood    *string `json:"mood"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid body")
			return
		}
		patch := store.JournalPatch{Title: body.Title, Content: body.Content}
		if body.Mood != nil {
			patch.Mood = &body.Mood
		}
		e, err := js.Update(id, patch)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, e)
	})

	r.Delete("/journal/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := js.Delete(id); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
}
