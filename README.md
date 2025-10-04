# MI6 Agent Manager

The **MI6 Agent Manager** is a lightweight, self-contained mock server management utility built with Go. It allows developers to configure, start, and stop multiple independent mock HTTP servers (Agents) from a single centralized web interface. It uses **Go**, **SQLite**, **HTMX**, **a-h/templ**, and **Tailwind CSS** for a fast, modern "backend-driven" user experience.

## 🚀 Features

  * **Mock Server Registry:** Centralized management (create, list, manage) of multiple Agents.
  * **Dynamic Server Lifecycle:** Agents can be started and stopped on unique ports without interrupting the main MI6 server.
  * **Persistent Configuration:** All Agent and path configurations are stored in an SQLite database.
  * **Modern UI:** Live-reloading web interface built with HTMX and a-h/templ for instant user feedback.

## 🛠 Project Structure

The project follows a standard modular Go layout to separate concerns:

```
mi6-project/
├── cmd/mi6/          # Main application entry point (main.go)
├── internal/
│   ├── agent/        # Core business logic (Registry, server management)
│   ├── api/          # Main MI6 API routes and handlers (JSON/HTMX responses)
│   ├── db/           # Data layer interface and SQLite implementation
│   └── web/          # Web UI handlers (Dashboard, Fragments)
├── web/
│   ├── static/       # Compiled CSS and HTMX/JS
│   └── template/     # a-h/templ components (.templ files)
└── Makefile          # Build, setup, and development commands
```

## ⚙️ Prerequisites

1.  **Go (1.21+)**
2.  **Node.js & npm** (Required for Tailwind CSS)
3.  **Air** (Recommended for development)
    ```bash
    go install github.com/cosmtrek/air@latest
    ```

## 💻 Development Setup

The `Makefile` simplifies the entire setup and development process.

### Step 1: Initial Setup

Run the setup command to install all Go and Node dependencies.

```bash
make setup
```

### Step 2: Start Development

You need **two separate terminals** running for the full live-reload experience:

| Terminal Window | Command | Purpose |
| :--- | :--- | :--- |
| **Terminal 1** | `make dev` | Watches all `.go` and `.templ` files, rebuilds, and restarts the Go server using **Air**. |
| **Terminal 2** | `make css-watch` | Starts the separate Tailwind CLI watcher to recompile CSS on file changes. |

Your MI6 Agent Manager will be accessible at `http://localhost:6969`.

-----

## 🔧 Build & Utility Commands

| Command | Description |
| :--- | :--- |
| `make build` | Compiles the Go application (`generate` and `go build`). |
| `make run` | Compiles and runs the server directly (non-development mode). |
| `make generate` | Compiles `.templ` files into runnable Go source code. |
| `make css-build` | Compiles a minified, production-ready `output.css` file. |
| `make clean` | Removes all compiled files, temp directories (`tmp/`), and the SQLite database (`agents.db`). |

## 🖥 API Usage

The following endpoints are primarily used by the HTMX frontend but can be accessed directly:

### Create a New Agent

**Endpoint:** `POST /agents`
**Body (JSON):**

```json
{
    "name": "Service_A",
    "port": "8081",
    "paths": [
        {
            "path": "/health",
            "response": "{\"status\": \"ok\"}"
        },
        {
            "path": "/api/users/123",
            "response": "{\"user_id\": 123, \"name\": \"Jane Doe\"}"
        }
    ]
}
```

### Server Control Endpoints (Used by HTMX)

| Action | Endpoint | Method |
| :--- | :--- | :--- |
| **List Agents** | `/agents` | `GET` |
| **Start Agent** | `/agents/{agentID}/start` | `POST` |
| **Stop Agent** | `/agents/{agentID}/stop` | `POST` |

