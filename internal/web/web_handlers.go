package web

import (
	"log"
	"net/http"

	"mi6/internal/db"
	"mi6/web/template" // Import your templ components
)

type Handlers struct {
	Repo db.AgentRepository
}

func NewHandlers(repo db.AgentRepository) *Handlers {
	return &Handlers{Repo: repo}
}

// DashboardPage handles the initial page load (full HTML render)
func (h *Handlers) DashboardPage(w http.ResponseWriter, r *http.Request) {
	agents, err := h.Repo.ListAgents(r.Context())
	if err != nil {
		http.Error(w, "Failed to load agents", http.StatusInternalServerError)
		return
	}

	// Render the main layout, injecting the list of agents
	template.Dashboard(agents).Render(r.Context(), w)
}

// AgentTableFragment handles HTMX requests to refresh the agent table (partial HTML render)
func (h *Handlers) AgentTableFragment(w http.ResponseWriter, r *http.Request) {
	agents, err := h.Repo.ListAgents(r.Context())
	if err != nil {
		log.Println("Error listing agents for fragment:", err)
		http.Error(w, "Failed to refresh agents", http.StatusInternalServerError)
		return
	}

    // Render ONLY the table component
	template.AgentTable(agents).Render(r.Context(), w)
}
