package db

import (
	"context"
	"database/sql"
	"fmt"
)

// SQLiteRepository implements the AgentRepository interface using SQLite.
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new repository instance.
func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// GetAgentByID fetches a single agent by ID.
func (r *SQLiteRepository) GetAgentByID(ctx context.Context, id int) (*Agent, error) {
	var agent Agent
	row := r.db.QueryRowContext(ctx, "SELECT id, name, port, status FROM agents WHERE id = ?", id)
	if err := row.Scan(&agent.Id, &agent.Name, &agent.Port, &agent.Status); err != nil {
		return nil, err // sql.ErrNoRows if not found
	}
	return &agent, nil
}

// ListAgents fetches all agents.
func (r *SQLiteRepository) ListAgents(ctx context.Context) ([]Agent, error) {
	var agents []Agent
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, port, status FROM agents")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var agent Agent
		if err := rows.Scan(&agent.Id, &agent.Name, &agent.Port, &agent.Status); err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}
	return agents, nil
}

// CreateAgent handles both the agent and its associated paths in a transaction.
func (r *SQLiteRepository) CreateAgent(ctx context.Context, name, port string, paths []AgentPath) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	// 1. Insert Agent
	res, err := tx.Exec("INSERT INTO agents(name, port, status) VALUES (?, ?, ?)", name, port, "stopped")
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to insert agent: %w", err)
	}
	agentID, _ := res.LastInsertId()

	// 2. Insert Agent Paths
	for _, p := range paths {
		_, err := tx.Exec("INSERT INTO agent_paths(agent_id, path, response) VALUES (?, ?, ?)", agentID, p.Path, p.Response)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to insert agent path: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("transaction commit failed: %w", err)
	}

	return int(agentID), nil
}

// UpdateAgentStatus updates the status of a specific agent.
func (r *SQLiteRepository) UpdateAgentStatus(ctx context.Context, id int, status string) error {
	stmt, err := r.db.Prepare("UPDATE agents SET status = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare update: %w", err)
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, status, id); err != nil {
		return fmt.Errorf("failed to execute status update: %w", err)
	}
	return nil
}

// GetAgentPaths fetches all paths associated with a given agent ID.
func (r *SQLiteRepository) GetAgentPaths(ctx context.Context, agentID int) ([]AgentPath, error) {
	var paths []AgentPath
	rows, err := r.db.QueryContext(ctx, "SELECT id, path, response FROM agent_paths WHERE agent_id = ?", agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p AgentPath
		if err := rows.Scan(&p.Id, &p.Path, &p.Response); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, nil
}
