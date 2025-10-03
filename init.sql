-- Agents Table: Holds configuration for each mock server (the Agent)
CREATE TABLE IF NOT EXISTS agents (
    id INTEGER PRIMARY KEY AUTOINCREMENT, -- Use AUTOINCREMENT for easy creation
    name TEXT NOT NULL UNIQUE,          -- Agent names should be unique
    port TEXT NOT NULL UNIQUE,          -- Ports must be unique to avoid collisions
    status TEXT NOT NULL DEFAULT 'stopped' -- Default status is helpful
);

-- Agent Paths Table: Holds the mock endpoints and their responses for an Agent
CREATE TABLE IF NOT EXISTS agent_paths (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id INTEGER NOT NULL,      -- CRITICAL: Foreign Key linking to the agents table
    path TEXT NOT NULL,
    response TEXT NOT NULL,

    -- Constraint: An agent cannot have the same path defined twice
    UNIQUE (agent_id, path),

    -- Define the Foreign Key relationship
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);
