package api

import (
	"github.com/go-chi/chi/v5"

	"gtzy/internal/store"
)

func (s *Server) registerTimerRoutes(r chi.Router) {}

func (s *Server) registerTimerActionRoutes(r chi.Router, ts *store.TaskStore) {}
