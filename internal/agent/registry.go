package agent

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"mi6/internal/db"

	"github.com/go-chi/chi/v5"
)

// Registry manages the lifecycle and access to all running mock servers.
type Registry struct {
	Servers map[int]*http.Server
	mu      sync.Mutex // Protects access to the Servers map
	Repo    db.AgentRepository
}

// NewRegistry creates a new agent registry instance.
func NewRegistry(repo db.AgentRepository) *Registry {
	return &Registry{
		Servers: make(map[int]*http.Server),
		Repo:    repo,
	}
}

// StartAgentServer retrieves configuration and launches the Agent server in a new goroutine.
func (r *Registry) StartAgentServer(ctx context.Context, agentID int) error {
	r.mu.Lock()
	if _, running := r.Servers[agentID]; running {
		r.mu.Unlock()
		return fmt.Errorf("agent %d is already running", agentID)
	}
	r.mu.Unlock()

	// 1. Get Agent details
	agent, err := r.Repo.GetAgentByID(ctx, agentID)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	// 2. Load paths from DB
	paths, err := r.Repo.GetAgentPaths(ctx, agentID)
	if err != nil {
		return fmt.Errorf("failed to load agent paths: %w", err)
	}

	// 3. Setup mock server router
	mux := chi.NewRouter()
	for _, p := range paths {
		path := p.Path // Capture loop variable
		response := p.Response
		mux.Get(path, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(response))
		})
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", agent.Port),
		Handler: mux,
	}

	// 4. Register and start in a new goroutine
	r.mu.Lock()
	r.Servers[agentID] = server
	r.mu.Unlock()

	go func() {
		log.Printf("Agent %d (%s) starting on port %s", agent.Id, agent.Name, agent.Port)

		// Update status to active
		r.Repo.UpdateAgentStatus(context.Background(), agent.Id, "active")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Agent %d failed to start: %v", agent.Id, err)
		}

		// Cleanup: Remove server from map and update status
		r.mu.Lock()
		delete(r.Servers, agent.Id)
		r.mu.Unlock()
		r.Repo.UpdateAgentStatus(context.Background(), agent.Id, "stopped")
		log.Printf("Agent %d stopped.", agent.Id)
	}()

	return nil
}

// StopAgentServer sends a graceful shutdown signal to a running agent.
func (r *Registry) StopAgentServer(agentID int) error {
	r.mu.Lock()
	server, running := r.Servers[agentID]
	r.mu.Unlock()

	if !running {
		return fmt.Errorf("agent %d is not running", agentID)
	}

	// Use a context for shutdown timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("agent %d shutdown failed: %w", agentID, err)
	}

	return nil
}

// ShutdownAll gracefully shuts down all running agents during application exit.
func (r *Registry) ShutdownAll() {
	r.mu.Lock()
	servers := make(map[int]*http.Server)
	for id, srv := range r.Servers {
		servers[id] = srv // Copy map to release lock quickly
	}
	r.mu.Unlock()

	if len(servers) == 0 {
		log.Println("No agents were running to shut down.")
		return
	}

	log.Printf("Shutting down %d active agents...", len(servers))
	var wg sync.WaitGroup

	for id, srv := range servers {
		wg.Add(1)
		go func(id int, srv *http.Server) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				log.Printf("Agent %d forced shutdown: %v", id, err)
			}
		}(id, srv)
	}

	wg.Wait()
	log.Println("All MI6 agents stopped.")
}
