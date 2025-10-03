package api

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"

	"mi6/internal/agent"
	"mi6/internal/db"

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
}

// NewRouter sets up the Chi router and API routes.
func NewRouter(mgr *agent.Registry) http.Handler {
	// Handlers struct holds the registry and repository dependencies
	h := &Handlers{
		Mgr:  mgr,
		Repo: mgr.Repo,
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// API routes for agent management
	r.Route("/agents", func(r chi.Router) {
		r.Get("/", h.ListAgents)
		r.Post("/", h.CreateAgent)
		// r.Post("/startAll", h.StartAllAgents) // To be implemented later

		r.Route("/{agentID}", func(r chi.Router) {
			r.Use(h.AgentCtx) // Middleware to load agent by ID
			r.Get("/", h.GetAgent)
			r.Post("/start", h.StartAgent)
			r.Post("/stop", h.StopAgent)
		})
	})

	return r
}

// AgentCtx is middleware that loads an Agent by ID and stores it in context.
func (h *Handlers) AgentCtx(next http.Handler) http.Handler {
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
