package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Agent represents a mock server configuration stored in the DB.
type Agent struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Port   string `json:"port"`
	Status string `json:"status"` // e.g., "stopped", "active"
}

// AgentPath defines a mock path and its response.
type AgentPath struct {
	Id       int
	AgentID  int
	Path     string `json:"path"`
	Response string `json:"response"`
}

// AgentRepository defines the interface for data access operations.
type AgentRepository interface {
	GetAgentByID(ctx context.Context, id int) (*Agent, error)
	ListAgents(ctx context.Context) ([]Agent, error)
	CreateAgent(ctx context.Context, name, port string, paths []AgentPath) (int, error)
	UpdateAgentStatus(ctx context.Context, id int, status string) error
	GetAgentPaths(ctx context.Context, agentID int) ([]AgentPath, error)
}

// --- Migrations ---

// RunMigrations initializes the SQLite schema.
func RunMigrations(db *sql.DB) error {
	const createAgentTable = `
	CREATE TABLE IF NOT EXISTS agents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		port TEXT NOT NULL UNIQUE,
		status TEXT NOT NULL DEFAULT 'stopped'
	);`

	const createAgentPathTable = `
	CREATE TABLE IF NOT EXISTS agent_paths (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		agent_id INTEGER NOT NULL,
		path TEXT NOT NULL,
		response TEXT NOT NULL,
		UNIQUE (agent_id, path),
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	);`

	if _, err := db.Exec(createAgentTable); err != nil {
		return fmt.Errorf("failed to create agents table: %w", err)
	}
	if _, err := db.Exec(createAgentPathTable); err != nil {
		return fmt.Errorf("failed to create agent_paths table: %w", err)
	}
	return nil
}
