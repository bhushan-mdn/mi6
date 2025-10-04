# ==============================================================================
# MI6 Agent Manager Makefile
# ==============================================================================

# --- Variables ---
BIN_NAME = mi6-manager
BUILD_DIR = ./tmp
RUNNER_PATH = $(BUILD_DIR)/$(BIN_NAME)

GO_CMD = go
NPM_CMD = npm
NPX_CMD = npx
AIR_CMD = air
TAILWIND_CLI = $(NPX_CMD) @tailwindcss/cli
TAILWIND_INPUT = ./web/static/input.css
TAILWIND_OUTPUT = ./web/static/output.css

.PHONY: setup clean build run dev db-migrate generate

# --- Default Target ---
all: setup build

# --- Setup and Cleaning ---

setup:
	@echo "--- 🛠 Setting up dependencies..."
	@$(GO_CMD) mod tidy
	@$(NPM_CMD) install

clean:
	@echo "--- 🧹 Cleaning build artifacts and data..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BIN_NAME)
	@rm -f *.db *.db-journal *.db-wal *.db-shm
	@rm -f $(TAILWIND_OUTPUT)
	@$(GO_CMD) clean -cache -modcache

# --- Build Targets ---

build: generate
	@echo "--- 🔨 Building Go application..."
	@mkdir -p $(BUILD_DIR)
	@$(GO_CMD) build -o $(RUNNER_PATH) ./cmd/mi6

# Compiles templ files into Go code
generate:
	@echo "--- ✨ Generating templ components..."
	@$(GO_CMD) generate ./...

# Compiles Tailwind CSS
css-build:
	@echo "--- 🎨 Compiling Tailwind CSS..."
	@$(TAILWIND_CLI) -i $(TAILWIND_INPUT) -o $(TAILWIND_OUTPUT) --minify

# --- Runtime / Development Targets ---

# Runs the compiled binary (requires 'build' first)
run: build css-build
	@echo "--- 🚀 Starting server on port 6969..."
	@$(RUNNER_PATH) -port 6969

# Development command using Air (requires .air.toml)
dev:
	@echo "--- 💻 Starting Air for Go/Templ live reload..."
	@echo "    (CSS watcher must run in a separate terminal: make css-watch)"
	$(AIR_CMD)

# CSS watcher for development (must run separately from Air)
css-watch:
	@echo "--- 🎨 Starting Tailwind CSS watcher..."
	$(TAILWIND_CLI) -i $(TAILWIND_INPUT) -o $(TAILWIND_OUTPUT) --watch

# --- Database Targets ---

# Runs database migrations (creates/updates tables)
db-migrate:
	@echo "--- 💾 Running database migrations..."
	@$(GO_CMD) run ./cmd/mi6/main.go --migrate-only # Assumes you add a --migrate-only flag to main.go
