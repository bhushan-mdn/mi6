package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"mi6/internal/db"
	"mi6/web/template"
)

// --- Request/Response DTOs ---

// NewAgentRequest structure for POST /agents
type NewAgentRequest struct {
	Name string `json:"name"`
	Port string `json:"port"`
	Paths []db.AgentPath `json:"paths"`
}

// --- Handlers ---

func (h *Handlers) ListAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := h.Repo.ListAgents(r.Context())
	if err != nil {
		http.Error(w, "Error listing agents", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(agents); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (h *Handlers) GetAgent(w http.ResponseWriter, r *http.Request) {
	// Agent is already in the context from the AgentCtx middleware
	agent, ok := r.Context().Value(keyAgent).(*db.Agent)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(agent); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (h *Handlers) CreateAgent(w http.ResponseWriter, r *http.Request) {
	var req NewAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// 1. Validate paths have AgentID 0, since it's a new agent
	for i := range req.Paths {
		req.Paths[i].AgentID = 0
	}

	// 2. Create Agent and Paths via Repository
	agentID, err := h.Repo.CreateAgent(r.Context(), req.Name, req.Port, req.Paths)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating agent: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": agentID,
		"status": "created",
		"message": fmt.Sprintf("Agent %d created successfully.", agentID),
	})
}

// func (h *Handlers) StartAgent(w http.ResponseWriter, r *http.Request) {
// 	agent, ok := r.Context().Value(keyAgent).(*db.Agent)
// 	if !ok {
// 		http.Error(w, http.StatusText(422), 422)
// 		return
// 	}

// 	if err := h.Mgr.StartAgentServer(r.Context(), agent.Id); err != nil {
// 		http.Error(w, err.Error(), http.StatusConflict) // Use StatusConflict for "already running"
// 		return
// 	}

// 	w.WriteHeader(http.StatusAccepted)
// 	w.Write([]byte(fmt.Sprintf(`{"message":"Agent %d started on port %s"}`, agent.Id, agent.Port)))
// }

// func (h *Handlers) StopAgent(w http.ResponseWriter, r *http.Request) {
// 	agent, ok := r.Context().Value(keyAgent).(*db.Agent)
// 	if !ok {
// 		http.Error(w, http.StatusText(422), 422)
// 		return
// 	}

// 	if err := h.Mgr.StopAgentServer(agent.Id); err != nil {
// 		http.Error(w, err.Error(), http.StatusNotFound)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(fmt.Sprintf(`{"message":"Agent %d stopping..."}`, agent.Id)))
// }

// ... (imports and structs remain the same) ...

// ... (ListAgents, GetAgent, CreateAgent remain JSON) ...

func (h *Handlers) StartAgent(w http.ResponseWriter, r *http.Request) {
	agent, ok := r.Context().Value(keyAgent).(*db.Agent)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	if err := h.Mgr.StartAgentServer(r.Context(), agent.Id); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

    // REFRESH the agent data from DB after starting
    updatedAgent, err := h.Repo.GetAgentByID(r.Context(), agent.Id)
    if err != nil {
        http.Error(w, "Agent started but failed to refresh data.", http.StatusInternalServerError)
        return
    }

    // NEW: Render and return the HTML fragment for the single row
    w.WriteHeader(http.StatusOK)
    template.AgentRow(updatedAgent).Render(r.Context(), w)
}

func (h *Handlers) StopAgent(w http.ResponseWriter, r *http.Request) {
	agent, ok := r.Context().Value(keyAgent).(*db.Agent)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	if err := h.Mgr.StopAgentServer(agent.Id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

    // REFRESH the agent data from DB after stopping (status will be 'stopped')
    updatedAgent, err := h.Repo.GetAgentByID(r.Context(), agent.Id)
    if err != nil {
        http.Error(w, "Agent stopped but failed to refresh data.", http.StatusInternalServerError)
        return
    }

    // NEW: Render and return the HTML fragment for the single row
    w.WriteHeader(http.StatusOK)
    template.AgentRow(updatedAgent).Render(r.Context(), w)
}
