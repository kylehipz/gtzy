package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"gtzy/internal/store"
	"gtzy/internal/timer"
)

func (s *Server) registerTimerRoutes(r chi.Router) {
	svc := timer.New(s.DB)

	r.Get("/timer/current", func(w http.ResponseWriter, r *http.Request) {
		t, err := svc.Current()
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		if t == nil {
			writeJSON(w, http.StatusOK, map[string]any{"current": nil})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"current": t})
	})

	r.Post("/timer/next", func(w http.ResponseWriter, r *http.Request) {
		t, err := svc.Next()
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"current": t})
	})

	r.Post("/timer/pause", func(w http.ResponseWriter, r *http.Request) {
		t, err := svc.PauseCurrent()
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"current": t})
	})
}

// registerTimerActionRoutes wires the per-task /tasks/:id/start|pause|complete
// actions. Registered from tasks.go so it shares the /tasks/{id} route tree.
func (s *Server) registerTimerActionRoutes(r chi.Router, ts *store.TaskStore) {
	svc := timer.New(s.DB)

	r.Post("/tasks/{id}/start", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid id")
			return
		}
		t, err := svc.Start(id)
		if err == timer.ErrNotFound {
			writeErr(w, http.StatusNotFound, "task not found")
			return
		}
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, t)
	})

	r.Post("/tasks/{id}/pause", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid id")
			return
		}
		t, err := svc.Pause(id)
		if err == timer.ErrNotFound {
			writeErr(w, http.StatusNotFound, "task not found")
			return
		}
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, t)
	})

	r.Post("/tasks/{id}/complete", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid id")
			return
		}
		t, err := svc.Complete(id)
		if err == timer.ErrNotFound {
			writeErr(w, http.StatusNotFound, "task not found")
			return
		}
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, t)
	})
}
