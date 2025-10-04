package api

import (
	"context"
	"database/sql"
	"net/http"
	"path/filepath"
	"strconv"

	"mi6/internal/agent"
	"mi6/internal/db"
	"mi6/internal/web" // NEW: Import the web handlers

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// AgentContextKey is used to store Agent data in the request context.
type AgentContextKey string

const keyAgent AgentContextKey = "agent"

// Handlers contains the dependencies needed by all API handlers.
type Handlers struct {
	Mgr  *agent.Registry
	Repo db.AgentRepository
    Web *web.Handlers // NEW: Web Handlers dependency
}

// NewRouter sets up the Chi router and API routes.
func NewRouter(mgr *agent.Registry) http.Handler {
	repo := mgr.Repo

	// Instantiate Web Handlers
	webH := web.NewHandlers(repo)

	// Handlers struct holds all dependencies
	h := &Handlers{
		Mgr:  mgr,
		Repo: repo,
        Web:  webH,
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

    // 1. Static Files (CSS/JS)
    fileServer(r, "/static", http.Dir(filepath.Join("web", "static")))

    // 2. Web UI Routes (Renders HTML pages/fragments)
    r.Get("/", h.Web.DashboardPage)
    r.Get("/ui/agents", h.Web.AgentTableFragment) // HTMX endpoint for table updates

	// 3. API Routes (Remain mostly JSON, but Start/Stop return HTML fragments for HTMX)
	r.Route("/agents", func(r chi.Router) {
		r.Get("/", h.ListAgents)
		r.Post("/", h.CreateAgent)

		r.Route("/{agentID}", func(r chi.Router) {
			r.Use(h.AgentCtx)
			r.Get("/", h.GetAgent)
            // THESE NOW RETURN HTML FRAGMENTS
			r.Post("/start", h.StartAgent)
			r.Post("/stop", h.StopAgent)
		})
	})

	return r
}

// fileServer serves files from the specified folder.
func fileServer(r chi.Router, path string, root http.FileSystem) {
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	r.Handle(path+"*", http.StripPrefix(path, http.FileServer(root)))
}

// AgentCtx remains the same as before
func (h *Handlers) AgentCtx(next http.Handler) http.Handler {
    // ... (Your existing AgentCtx logic) ...
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agentIDStr := chi.URLParam(r, "agentID")
		agentID, err := strconv.Atoi(agentIDStr)
		if err != nil {
			http.Error(w, "Invalid Agent ID", http.StatusBadRequest)
			return
		}

		agent, err := h.Repo.GetAgentByID(r.Context(), agentID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, http.StatusText(404), 404)
				return
			}
			http.Error(w, http.StatusText(500), 500)
			return
		}

		ctx := context.WithValue(r.Context(), keyAgent, agent)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
