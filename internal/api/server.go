package api

import (
	"database/sql"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server holds shared dependencies for HTTP handlers.
type Server struct {
	DB *sql.DB
	AI AISummarizer
}

// AISummarizer is the subset of internal/ai.Client used by the summary handlers.
type AISummarizer interface {
	Enabled() bool
	Model() string
	Generate(periodType, periodKey, prompt string) (string, error)
}

// NewRouter builds the full chi router: /api/* endpoints plus, if spaFS is
// non-nil, the embedded SPA served for all other routes with index.html
// fallback.
func NewRouter(s *Server, spaFS fs.FS) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		})
		s.registerTaskRoutes(r)
		s.registerCategoryRoutes(r)
		s.registerRecurrenceRoutes(r)
		s.registerTimerRoutes(r)
		s.registerJournalRoutes(r)
		s.registerCalendarRoutes(r)
		s.registerStatsRoutes(r)
		s.registerBloodSugarRoutes(r)
		s.registerSummaryRoutes(r)
	})

	if spaFS != nil {
		fileServer := http.FileServer(http.FS(spaFS))
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			if _, err := fs.Stat(spaFS, "index.html"); err == nil {
				if _, statErr := fs.Stat(spaFS, r.URL.Path[1:]); statErr != nil {
					r.URL.Path = "/"
				}
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	return r
}
